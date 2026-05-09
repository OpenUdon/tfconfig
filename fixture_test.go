package tfconfig

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateFixtureGoldens = flag.Bool("update-fixtures", false, "update parser fixture expected JSON files")

func TestFixtureCorpus(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("testdata", "fixtures", "*", "input"))
	if err != nil {
		t.Fatalf("glob fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("no parser fixtures found")
	}

	for _, inputDir := range fixtures {
		inputDir := normalizePath(inputDir)
		name := filepath.Base(filepath.Dir(inputDir))
		t.Run(name, func(t *testing.T) {
			doc, err := LoadDir(inputDir)
			if err != nil {
				t.Fatalf("LoadDir(%s) failed: %v", inputDir, err)
			}
			got, err := doc.JSONIndent("", "  ")
			if err != nil {
				t.Fatalf("JSON projection failed: %v", err)
			}
			got = append(got, '\n')

			expectedPath := filepath.Join(filepath.Dir(inputDir), "expected.json")
			if *updateFixtureGoldens {
				if err := os.WriteFile(expectedPath, got, 0o644); err != nil {
					t.Fatalf("update %s: %v", expectedPath, err)
				}
				return
			}

			want, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read %s: %v", expectedPath, err)
			}
			if string(got) != string(want) {
				t.Fatalf("fixture output mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}
		})
	}
}

func TestFixtureCorpusCoversM6Scenarios(t *testing.T) {
	want := []string{
		"core-language",
		"diagnostics",
		"formats-overrides",
		"module-tree",
	}
	for _, name := range want {
		inputDir := filepath.Join("testdata", "fixtures", name, "input")
		if _, err := os.Stat(inputDir); err != nil {
			t.Fatalf("missing M6 fixture %s: %v", name, err)
		}
		expectedPath := filepath.Join("testdata", "fixtures", name, "expected.json")
		if _, err := os.Stat(expectedPath); err != nil && !*updateFixtureGoldens {
			t.Fatalf("missing M6 fixture golden %s: %v", expectedPath, err)
		}
	}
}

func TestOpenTofuEquivalenceFixtureCorpus(t *testing.T) {
	root := os.Getenv("OPENTOFU_EQUIVALENCE_TESTS")
	if root == "" {
		root = filepath.Join("..", "opentofu", "testing", "equivalence-tests", "tests")
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Skipf("OpenTofu equivalence fixture corpus not available at %s", root)
	}

	fixtures, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		t.Fatalf("glob OpenTofu equivalence fixtures: %v", err)
	}
	if len(fixtures) != 49 {
		t.Fatalf("OpenTofu equivalence fixture count = %d, want 49", len(fixtures))
	}

	for _, fixtureDir := range fixtures {
		fixtureDir := fixtureDir
		info, err := os.Stat(fixtureDir)
		if err != nil {
			t.Fatalf("stat %s: %v", fixtureDir, err)
		}
		if !info.IsDir() {
			continue
		}
		t.Run(filepath.Base(fixtureDir), func(t *testing.T) {
			doc, err := LoadDir(fixtureDir)
			if err != nil {
				t.Fatalf("LoadDir(%s) failed: %v", fixtureDir, err)
			}
			if errors := documentErrorDiagnostics(doc); len(errors) != 0 {
				t.Fatalf("LoadDir(%s) returned error diagnostics: %#v", fixtureDir, errors)
			}
		})
	}
}

func TestOpenTofuValidModulesFixtureCorpus(t *testing.T) {
	root := os.Getenv("OPENTOFU_VALID_MODULES")
	if root == "" {
		root = filepath.Join("..", "opentofu", "internal", "configs", "testdata", "valid-modules")
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Skipf("OpenTofu valid-modules fixture corpus not available at %s", root)
	}

	fixtures, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		t.Fatalf("glob OpenTofu valid-modules fixtures: %v", err)
	}
	if len(fixtures) != 34 {
		t.Fatalf("OpenTofu valid-modules fixture count = %d, want 34", len(fixtures))
	}
	expectedErrorCodes := map[string][]string{
		// This upstream config loader fixture is valid without loading child
		// modules. tfconfig intentionally attempts local module loading and
		// reports the missing fixture directory as a static diagnostic.
		"override-module": {"module_source_missing"},
	}

	for _, fixtureDir := range fixtures {
		fixtureDir := fixtureDir
		info, err := os.Stat(fixtureDir)
		if err != nil {
			t.Fatalf("stat %s: %v", fixtureDir, err)
		}
		if !info.IsDir() {
			continue
		}
		t.Run(filepath.Base(fixtureDir), func(t *testing.T) {
			doc, err := LoadDir(fixtureDir)
			if err != nil {
				t.Fatalf("LoadDir(%s) failed: %v", fixtureDir, err)
			}
			name := filepath.Base(fixtureDir)
			errors := documentErrorDiagnostics(doc)
			if wantCodes, ok := expectedErrorCodes[name]; ok {
				if !diagnosticCodesEqual(errors, wantCodes) {
					t.Fatalf("LoadDir(%s) error diagnostics = %#v, want codes %v", fixtureDir, errors, wantCodes)
				}
				return
			}
			if len(errors) != 0 {
				t.Fatalf("LoadDir(%s) returned error diagnostics: %#v", fixtureDir, errors)
			}
		})
	}
}

func TestOpenTofuValidModulesFixtureSemantics(t *testing.T) {
	root := os.Getenv("OPENTOFU_VALID_MODULES")
	if root == "" {
		root = filepath.Join("..", "opentofu", "internal", "configs", "testdata", "valid-modules")
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		t.Skipf("OpenTofu valid-modules fixture corpus not available at %s", root)
	}

	t.Run("backend and cloud overrides", func(t *testing.T) {
		backend := loadOpenTofuValidModuleFixture(t, root, "override-backend")
		backendMod := requireModule(t, backend, "")
		if len(backendMod.Backends) != 1 || backendMod.Backends[0].Type != "bar" {
			t.Fatalf("override-backend backends = %#v, want only bar", backendMod.Backends)
		}
		if got, ok := attributeLiteralString(backendMod.Backends[0].Config, "path"); !ok || got != "CHANGED/relative/path/to/terraform.tfstate" {
			t.Fatalf("override-backend path = %q ok=%v", got, ok)
		}

		cloud := loadOpenTofuValidModuleFixture(t, root, "override-cloud")
		cloudMod := requireModule(t, cloud, "")
		if cloudMod.Cloud == nil {
			t.Fatalf("override-cloud did not decode cloud config")
		}
		if got, ok := attributeLiteralString(cloudMod.Cloud.Config, "organization"); !ok || got != "CHANGED" {
			t.Fatalf("override-cloud organization = %q ok=%v", got, ok)
		}
		if _, ok := attributeLiteralString(cloudMod.Cloud.Config, "should_not_be_present_with_override"); ok {
			t.Fatalf("override-cloud preserved overridden attribute: %#v", cloudMod.Cloud.Config)
		}

		cloudOverBackend := loadOpenTofuValidModuleFixture(t, root, "override-backend-with-cloud")
		cloudOverBackendMod := requireModule(t, cloudOverBackend, "")
		if len(cloudOverBackendMod.Backends) != 0 {
			t.Fatalf("cloud override should clear backend: %#v", cloudOverBackendMod.Backends)
		}
		if cloudOverBackendMod.Cloud == nil {
			t.Fatalf("cloud override did not decode cloud config")
		}
	})

	t.Run("provider meta and ephemeral resources", func(t *testing.T) {
		providerMeta := loadOpenTofuValidModuleFixture(t, root, "provider-meta")
		metaMod := requireModule(t, providerMeta, "")
		if len(metaMod.ProviderMetas) != 1 || metaMod.ProviderMetas[0].Provider != "my-provider" {
			t.Fatalf("provider meta = %#v, want my-provider", metaMod.ProviderMetas)
		}
		if got, ok := attributeLiteralString(metaMod.ProviderMetas[0].Config, "hello"); !ok || got != "test-module" {
			t.Fatalf("provider meta hello = %q ok=%v", got, ok)
		}

		ephemeral := loadOpenTofuValidModuleFixture(t, root, "nested-providers-fqns")
		rootMod := requireModule(t, ephemeral, "")
		eph, ok := ephemeralResourceByAddress(rootMod.EphemeralResources, "ephemeral.test_ephemeral.explicit")
		if !ok {
			t.Fatalf("ephemeral resource not decoded: %#v", rootMod.EphemeralResources)
		}
		if eph.Provider == nil || eph.Provider.Address != "provider.foo-test" {
			t.Fatalf("ephemeral provider = %#v, want provider.foo-test", eph.Provider)
		}
	})

	t.Run("test file variables and plan options", func(t *testing.T) {
		doc := loadOpenTofuValidModuleFixture(t, root, "with-tests")
		mod := requireModule(t, doc, "")
		testFile, ok := testFileByPath(mod.Tests, "test_case_one.tftest.hcl")
		if !ok {
			t.Fatalf("test_case_one.tftest.hcl not decoded: %#v", mod.Tests)
		}
		if got, ok := attributeLiteralString(testFile.Variables, "input"); !ok || got != "default" {
			t.Fatalf("top-level test variable input = %q ok=%v", got, ok)
		}
		run, ok := testRunByName(testFile.Runs, "test_run_one")
		if !ok {
			t.Fatalf("test_run_one not decoded: %#v", testFile.Runs)
		}
		if run.Command != "plan" || len(run.PlanOptions) == 0 || len(run.Assertions) == 0 {
			t.Fatalf("test run facts not decoded: %#v", run)
		}
	})
}

func loadOpenTofuValidModuleFixture(t *testing.T, root, name string) Document {
	t.Helper()
	doc, err := LoadDir(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("LoadDir(%s) failed: %v", name, err)
	}
	if errors := documentErrorDiagnostics(doc); len(errors) != 0 {
		t.Fatalf("LoadDir(%s) returned error diagnostics: %#v", name, errors)
	}
	return doc
}

func ephemeralResourceByAddress(resources []EphemeralResource, address string) (EphemeralResource, bool) {
	for _, resource := range resources {
		if resource.Address == address {
			return resource, true
		}
	}
	return EphemeralResource{}, false
}

func testFileByPath(files []TestFile, path string) (TestFile, bool) {
	for _, file := range files {
		if file.Path == path {
			return file, true
		}
	}
	return TestFile{}, false
}

func testRunByName(runs []TestRun, name string) (TestRun, bool) {
	for _, run := range runs {
		if run.Name == name {
			return run, true
		}
	}
	return TestRun{}, false
}

func diagnosticCodesEqual(diags []Diagnostic, want []string) bool {
	if len(diags) != len(want) {
		return false
	}
	got := make([]string, len(diags))
	for i, diag := range diags {
		got[i] = diag.Code
	}
	return strings.Join(got, "\x00") == strings.Join(want, "\x00")
}

func TestFixtureGoldensDoNotLeakKnownSecrets(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("testdata", "fixtures", "*", "expected.json"))
	if err != nil {
		t.Fatalf("glob fixture goldens: %v", err)
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, secret := range []string{"plain-secret", "token-from-fixture", "inline-password"} {
			if strings.Contains(string(data), secret) {
				t.Fatalf("fixture golden %s leaked known secret literal %q", path, secret)
			}
		}
	}
}

func documentErrorDiagnostics(doc Document) []Diagnostic {
	var out []Diagnostic
	for _, diag := range doc.Diagnostics {
		if diag.Severity == DiagnosticError {
			out = append(out, diag)
		}
	}
	for _, mod := range doc.Modules {
		for _, diag := range mod.Diagnostics {
			if diag.Severity == DiagnosticError {
				out = append(out, diag)
			}
		}
	}
	return out
}
