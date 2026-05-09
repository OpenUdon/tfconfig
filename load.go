package tfconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
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
	doc.SourceRoots = append(doc.SourceRoots, SourceRoot{
		ID:            "root",
		ModuleAddress: "",
		Dir:           normalizePath(dir),
	})

	files, diags := discoverRootFiles(absDir, opts)
	doc.Diagnostics = append(doc.Diagnostics, diags...)
	for _, file := range files {
		doc.SourceFiles = append(doc.SourceFiles, SourceFile{
			ID:            file.RelPath,
			ModuleAddress: "",
			Path:          file.RelPath,
			Format:        file.Format,
			Role:          file.Role,
		})
	}

	mod := Module{
		Address:     "",
		Dir:         ".",
		Status:      ModuleStatusRoot,
		SourceFiles: sourceFilePaths(files, FileRolePrimary, FileRoleOverride),
	}

	parser := hclparse.NewParser()
	sources := make(map[string]sourceInfo)
	for _, file := range files {
		data, _ := os.ReadFile(file.AbsPath)
		sources[file.AbsPath] = sourceInfo{id: file.RelPath, path: file.RelPath, data: data}
	}

	for _, file := range files {
		body, parseDiags := parseFile(parser, file)
		doc.Diagnostics = append(doc.Diagnostics, convertDiagnostics(parseDiags, "", "", sources)...)
		if body == nil {
			continue
		}
		if file.Role == FileRoleTest {
			tests, testDiags := decodeTestFile(file, body, sources)
			mod.Diagnostics = append(mod.Diagnostics, testDiags...)
			mod.Tests = append(mod.Tests, tests...)
			continue
		}
		fileDiags := decodeConfigFile(&mod, file, body, sources)
		mod.Diagnostics = append(mod.Diagnostics, fileDiags...)
	}

	doc.Modules = append(doc.Modules, mod)
	return doc.Canonical(), nil
}

func discoverRootFiles(absDir string, opts LoadOptions) ([]discoveredFile, []Diagnostic) {
	var diags []Diagnostic
	var primary, override, tests []discoveredFile

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, []Diagnostic{{
			Severity: DiagnosticError,
			Code:     "read_module_directory",
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", absDir),
		}}
	}

	for _, entry := range entries {
		if entry.IsDir() || isIgnoredFile(entry.Name()) {
			continue
		}
		file, ok := classifyFile(absDir, absDir, entry.Name(), FileRolePrimary)
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
	testPath := filepath.Join(absDir, testDir)
	testEntries, err := os.ReadDir(testPath)
	if err == nil {
		for _, entry := range testEntries {
			if entry.IsDir() || isIgnoredFile(entry.Name()) {
				continue
			}
			file, ok := classifyFile(absDir, testPath, entry.Name(), FileRoleTest)
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
