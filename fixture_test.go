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
