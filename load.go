package tfconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

const defaultTestDirectory = "tests"

const (
	tfExt           = ".tf"
	tofuExt         = ".tofu"
	tfJSONExt       = ".tf.json"
	tofuJSONExt     = ".tofu.json"
	tfTestExt       = ".tftest.hcl"
	tofuTestExt     = ".tofutest.hcl"
	tfTestJSONExt   = ".tftest.json"
	tofuTestJSONExt = ".tofutest.json"
)

// LoadOptions controls optional static loader behavior.
type LoadOptions struct {
	// Producer overrides the default producer metadata when non-empty.
	Producer string

	// TestDir is the optional module test directory to scan. If empty, "tests"
	// is used. Test files directly in the root directory are always discovered.
	TestDir string
}

type discoveredFile struct {
	AbsPath string
	RelPath string
	Format  FileFormat
	Role    FileRole
}

// LoadDir loads a Terraform/OpenTofu configuration directory into the static v1
// fact model. Diagnostics are returned in the document; the error return is
// reserved for setup failures outside normal configuration diagnostics.
func LoadDir(dir string) (Document, error) {
	return LoadDirWithOptions(dir, LoadOptions{})
}

// LoadDirWithOptions loads a Terraform/OpenTofu configuration directory using
// opts and returns a canonical static v1 document.
func LoadDirWithOptions(dir string, opts LoadOptions) (Document, error) {
	if dir == "" {
		dir = "."
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return Document{}, err
	}

	doc := NewDocument(normalizePath(dir))
	if opts.Producer != "" {
		doc.Producer = opts.Producer
	}

	loader := moduleLoader{
		rootAbs: absDir,
		opts:    opts,
		parser:  hclparse.NewParser(),
		sources: make(map[string]sourceInfo),
	}
	stack := map[string]bool{}
	loader.loadModule(&doc, absDir, "", "", nil, ModuleStatusRoot, stack)
	return doc.Canonical(), nil
}

type moduleLoader struct {
	rootAbs string
	opts    LoadOptions
	parser  *hclparse.Parser
	sources map[string]sourceInfo
}

func (l *moduleLoader) loadModule(doc *Document, moduleAbs, address, parentAddress string, source *Value, status ModuleStatus, stack map[string]bool) {
	moduleAbs = filepath.Clean(moduleAbs)
	canonical := canonicalModuleDir(moduleAbs)
	stack[canonical] = true
	defer delete(stack, canonical)

	files, diags := discoverModuleFiles(l.rootAbs, moduleAbs, l.opts)
	moduleDiags := diags
	if address == "" {
		doc.Diagnostics = append(doc.Diagnostics, diags...)
		moduleDiags = nil
	} else {
		for i := range moduleDiags {
			moduleDiags[i].ModuleAddress = address
		}
	}

	mod := Module{
		Address:       address,
		Source:        cloneValuePtr(source),
		Dir:           moduleDir(l.rootAbs, moduleAbs, address),
		ParentAddress: parentAddress,
		Status:        status,
		SourceFiles:   sourceFilePaths(files, FileRolePrimary, FileRoleOverride),
		Diagnostics:   moduleDiags,
	}

	doc.SourceRoots = append(doc.SourceRoots, SourceRoot{
		ID:            sourceRootID(address),
		ModuleAddress: address,
		Dir:           sourceRootDir(doc.RootDir, mod.Dir, address),
	})
	for _, file := range files {
		doc.SourceFiles = append(doc.SourceFiles, SourceFile{
			ID:            file.RelPath,
			ModuleAddress: address,
			Path:          file.RelPath,
			Format:        file.Format,
			Role:          file.Role,
		})
		data, _ := os.ReadFile(file.AbsPath)
		l.sources[file.AbsPath] = sourceInfo{id: file.RelPath, path: file.RelPath, data: data}
	}

	for _, file := range files {
		body, parseDiags := parseFile(l.parser, file)
		doc.Diagnostics = append(doc.Diagnostics, convertDiagnostics(parseDiags, address, "", l.sources)...)
		if body == nil {
			continue
		}
		if file.Role == FileRoleTest {
			tests, testDiags := decodeTestFile(file, body, l.sources)
			mod.Diagnostics = append(mod.Diagnostics, testDiags...)
			mod.Tests = append(mod.Tests, tests...)
			continue
		}
		fileDiags := decodeConfigFile(&mod, file, body, l.sources)
		mod.Diagnostics = append(mod.Diagnostics, fileDiags...)
	}

	for i := range mod.ModuleCalls {
		mod.ModuleCalls[i].Address = childModuleAddress(address, mod.ModuleCalls[i].Name)
	}

	doc.Modules = append(doc.Modules, mod)
	for _, call := range sortedModuleCalls(mod.ModuleCalls) {
		childAddress := childModuleAddress(address, call.Name)
		resolution := classifyModuleSource(call, moduleAbs, l.rootAbs, childAddress)
		if resolution.status != ModuleStatusLoaded {
			doc.Modules = append(doc.Modules, placeholderModule(call, resolution, address, childAddress))
			continue
		}
		childCanonical := canonicalModuleDir(resolution.absDir)
		if stack[childCanonical] {
			cycleResolution := resolution
			cycleResolution.status = ModuleStatusUnsupported
			cycleResolution.diag = moduleSourceDiag(
				DiagnosticError,
				"module_source_cycle",
				"Local module source creates a cycle",
				fmt.Sprintf("Module %s resolves to %s, which is already on the active module loading stack.", childAddress, resolution.dir),
				childAddress,
				call.Address,
				sourceRangeForModuleSource(call),
			)
			doc.Modules = append(doc.Modules, placeholderModule(call, cycleResolution, address, childAddress))
			continue
		}
		l.loadModule(doc, resolution.absDir, childAddress, address, call.Source, ModuleStatusLoaded, stack)
	}
}

func sortedModuleCalls(calls []ModuleCall) []ModuleCall {
	out := cloneModuleCalls(calls)
	sort.Slice(out, func(i, j int) bool { return out[i].Address < out[j].Address })
	return out
}

func discoverModuleFiles(rootAbs, moduleAbs string, opts LoadOptions) ([]discoveredFile, []Diagnostic) {
	var diags []Diagnostic
	var primary, override, tests []discoveredFile

	entries, err := os.ReadDir(moduleAbs)
	if err != nil {
		return nil, []Diagnostic{{
			Severity: DiagnosticError,
			Code:     "read_module_directory",
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", moduleAbs),
		}}
	}

	for _, entry := range entries {
		if entry.IsDir() || isIgnoredFile(entry.Name()) {
			continue
		}
		file, ok := classifyFile(rootAbs, moduleAbs, entry.Name(), FileRolePrimary)
		if !ok {
			continue
		}
		if file.Role == FileRoleTest {
			tests = append(tests, file)
			continue
		}
		if isOverrideFile(entry.Name()) {
			file.Role = FileRoleOverride
			override = append(override, file)
		} else {
			primary = append(primary, file)
		}
	}

	testDir := opts.TestDir
	if testDir == "" {
		testDir = defaultTestDirectory
	}
	testPath := filepath.Join(moduleAbs, testDir)
	testEntries, err := os.ReadDir(testPath)
	if err == nil {
		for _, entry := range testEntries {
			if entry.IsDir() || isIgnoredFile(entry.Name()) {
				continue
			}
			file, ok := classifyFile(rootAbs, testPath, entry.Name(), FileRoleTest)
			if ok && file.Role == FileRoleTest {
				tests = append(tests, file)
			}
		}
	} else if !os.IsNotExist(err) {
		diags = append(diags, Diagnostic{
			Severity: DiagnosticWarning,
			Code:     "read_test_directory",
			Summary:  "Failed to read test directory",
			Detail:   fmt.Sprintf("Test directory %s could not be read: %v.", testPath, err),
		})
	}

	primary = filterTfFilesWithTofuAlternatives(primary)
	override = filterTfFilesWithTofuAlternatives(override)
	tests = filterTfFilesWithTofuAlternatives(tests)

	files := make([]discoveredFile, 0, len(primary)+len(override)+len(tests))
	files = append(files, primary...)
	files = append(files, override...)
	files = append(files, tests...)
	return files, diags
}

type moduleSourceResolution struct {
	status ModuleStatus
	absDir string
	dir    string
	diag   *Diagnostic
}

func classifyModuleSource(call ModuleCall, parentAbs, rootAbs, childAddress string) moduleSourceResolution {
	rng := sourceRangeForModuleSource(call)
	if call.Source == nil {
		return moduleSourceResolution{
			status: ModuleStatusUnsupported,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_unsupported",
				"Module source is not statically loadable",
				"Module source is missing, so no child module directory can be loaded statically.",
				childAddress,
				call.Address,
				rng,
			),
		}
	}
	source, ok := call.Source.Literal.(string)
	if !ok || call.Source.Kind != ValueKindString {
		return moduleSourceResolution{
			status: ModuleStatusUnsupported,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_unsupported",
				"Module source is not statically loadable",
				"Module source is an expression rather than a direct local filesystem path.",
				childAddress,
				call.Address,
				rng,
			),
		}
	}
	source = strings.TrimSpace(source)
	if source == "" {
		return moduleSourceResolution{
			status: ModuleStatusUnsupported,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_unsupported",
				"Module source is not statically loadable",
				"Module source is empty, so no child module directory can be loaded statically.",
				childAddress,
				call.Address,
				rng,
			),
		}
	}
	if !isDirectLocalModuleSource(source) {
		return moduleSourceResolution{
			status: ModuleStatusRemote,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_remote",
				"Module source requires external retrieval",
				fmt.Sprintf("Module source %q is registry, remote, or downloader-backed and was not loaded. The static loader does not run init or downloader steps.", source),
				childAddress,
				call.Address,
				rng,
			),
		}
	}

	absDir := source
	if !filepath.IsAbs(absDir) {
		absDir = filepath.Join(parentAbs, source)
	}
	absDir = filepath.Clean(absDir)
	dir := moduleDir(rootAbs, absDir, childAddress)
	info, err := os.Stat(absDir)
	if err != nil {
		return moduleSourceResolution{
			status: ModuleStatusMissing,
			absDir: absDir,
			dir:    dir,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_missing",
				"Local module source is not readable",
				fmt.Sprintf("Module source %q resolved to %s, which does not exist or cannot be read.", source, dir),
				childAddress,
				call.Address,
				rng,
			),
		}
	}
	if !info.IsDir() {
		return moduleSourceResolution{
			status: ModuleStatusMissing,
			absDir: absDir,
			dir:    dir,
			diag: moduleSourceDiag(
				DiagnosticError,
				"module_source_missing",
				"Local module source is not a directory",
				fmt.Sprintf("Module source %q resolved to %s, which is not a directory.", source, dir),
				childAddress,
				call.Address,
				rng,
			),
		}
	}
	return moduleSourceResolution{status: ModuleStatusLoaded, absDir: absDir, dir: dir}
}

func placeholderModule(call ModuleCall, resolution moduleSourceResolution, parentAddress, childAddress string) Module {
	mod := Module{
		Address:       childAddress,
		Source:        cloneValuePtr(call.Source),
		Dir:           resolution.dir,
		ParentAddress: parentAddress,
		Status:        resolution.status,
		Range:         cloneSourceRange(call.Range),
	}
	if resolution.diag != nil {
		mod.Diagnostics = append(mod.Diagnostics, *resolution.diag)
	}
	return mod
}

func moduleSourceDiag(severity DiagnosticSeverity, code, summary, detail, moduleAddress, address string, rng *SourceRange) *Diagnostic {
	diag := modelDiagnostic(severity, code, summary, detail, moduleAddress, address, rng)
	return &diag
}

func sourceRangeForModuleSource(call ModuleCall) *SourceRange {
	if call.Source != nil && call.Source.Range != nil {
		return cloneSourceRange(call.Source.Range)
	}
	return cloneSourceRange(call.Range)
}

func isDirectLocalModuleSource(source string) bool {
	return source == "." ||
		source == ".." ||
		strings.HasPrefix(source, "./") ||
		strings.HasPrefix(source, "../") ||
		filepath.IsAbs(source)
}

func moduleDir(rootAbs, moduleAbs, address string) string {
	if address == "" {
		return "."
	}
	if rel, err := filepath.Rel(rootAbs, moduleAbs); err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
		return normalizePath(rel)
	}
	return normalizePath(moduleAbs)
}

func sourceRootDir(rootDir, moduleDir, address string) string {
	if address == "" {
		return normalizePath(rootDir)
	}
	return moduleDir
}

func sourceRootID(address string) string {
	if address == "" {
		return "root"
	}
	return address
}

func childModuleAddress(parentAddress, name string) string {
	if parentAddress == "" {
		return "module." + name
	}
	return parentAddress + ".module." + name
}

func canonicalModuleDir(absDir string) string {
	resolved, err := filepath.EvalSymlinks(absDir)
	if err == nil {
		return filepath.Clean(resolved)
	}
	abs, err := filepath.Abs(absDir)
	if err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(absDir)
}

func classifyFile(rootDir, dir, name string, defaultRole FileRole) (discoveredFile, bool) {
	ext := fileExt(name)
	if ext == "" {
		return discoveredFile{}, false
	}

	format := FileFormatHCL
	if strings.HasSuffix(ext, ".json") {
		format = FileFormatJSON
	}
	role := defaultRole
	if isTestFileExt(ext) {
		role = FileRoleTest
	}
	absPath := filepath.Join(dir, name)
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		rel = name
	}
	return discoveredFile{
		AbsPath: absPath,
		RelPath: normalizePath(rel),
		Format:  format,
		Role:    role,
	}, true
}

func parseFile(parser *hclparse.Parser, file discoveredFile) (hcl.Body, hcl.Diagnostics) {
	var parsed *hcl.File
	var diags hcl.Diagnostics
	if file.Format == FileFormatJSON {
		parsed, diags = parser.ParseJSONFile(file.AbsPath)
	} else {
		parsed, diags = parser.ParseHCLFile(file.AbsPath)
	}
	if parsed == nil {
		return nil, diags
	}
	return parsed.Body, diags
}

func sourceFilePaths(files []discoveredFile, roles ...FileRole) []string {
	roleSet := map[FileRole]bool{}
	for _, role := range roles {
		roleSet[role] = true
	}
	var paths []string
	for _, file := range files {
		if roleSet[file.Role] {
			paths = append(paths, file.RelPath)
		}
	}
	return paths
}

func isOverrideFile(name string) bool {
	ext := fileExt(name)
	if ext == "" {
		return false
	}
	base := name[:len(name)-len(ext)]
	return base == "override" || strings.HasSuffix(base, "_override")
}

func fileExt(path string) string {
	if ext := tfFileExt(path); ext != "" {
		return ext
	}
	return tofuFileExt(path)
}

func tfFileExt(path string) string {
	switch {
	case strings.HasSuffix(path, tfJSONExt):
		return tfJSONExt
	case strings.HasSuffix(path, tfExt):
		return tfExt
	case strings.HasSuffix(path, tfTestJSONExt):
		return tfTestJSONExt
	case strings.HasSuffix(path, tfTestExt):
		return tfTestExt
	default:
		return ""
	}
}

func tofuFileExt(path string) string {
	switch {
	case strings.HasSuffix(path, tofuJSONExt):
		return tofuJSONExt
	case strings.HasSuffix(path, tofuExt):
		return tofuExt
	case strings.HasSuffix(path, tofuTestJSONExt):
		return tofuTestJSONExt
	case strings.HasSuffix(path, tofuTestExt):
		return tofuTestExt
	default:
		return ""
	}
}

func isTestFileExt(ext string) bool {
	return ext == tfTestExt || ext == tfTestJSONExt || ext == tofuTestExt || ext == tofuTestJSONExt
}

func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") ||
		strings.HasSuffix(name, "~") ||
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#")
}

func filterTfFilesWithTofuAlternatives(paths []discoveredFile) []discoveredFile {
	var ignored []string
	var relevant []discoveredFile
	all := make([]string, 0, len(paths))
	byAbs := make(map[string]discoveredFile, len(paths))
	for _, file := range paths {
		all = append(all, file.AbsPath)
		byAbs[file.AbsPath] = file
	}
	for _, file := range paths {
		ext := tfFileExt(file.AbsPath)
		if ext == "" {
			relevant = append(relevant, file)
			continue
		}
		parallelTofuExt := strings.ReplaceAll(ext, ".tf", ".tofu")
		pathWithoutExt, _ := strings.CutSuffix(file.AbsPath, ext)
		parallelTofuPath := pathWithoutExt + parallelTofuExt
		if slices.Contains(all, parallelTofuPath) {
			ignored = append(ignored, file.RelPath)
			continue
		}
		relevant = append(relevant, byAbs[file.AbsPath])
	}
	if len(ignored) > 0 {
		log.Printf("[INFO] ignored .tf files because .tofu alternatives exist: %q", ignored)
	}
	return relevant
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}
