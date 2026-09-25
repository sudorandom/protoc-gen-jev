package golden

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var Update = flag.Bool("update", false, "update .golden files")

// AssertMatches compares actual bytes against the golden file.
// If -update flag or UPDATE_GOLDEN=1 is set, it updates the golden file.
func AssertMatches(t *testing.T, goldenPath string, actualBytes []byte) {
	t.Helper()

	shouldUpdate := *Update || os.Getenv("UPDATE_GOLDEN") == "1" || os.Getenv("UPDATE_GOLDEN") == "true"

	if shouldUpdate {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
			t.Fatalf("failed to create directory for golden file %s: %v", goldenPath, err)
		}
		if err := os.WriteFile(goldenPath, actualBytes, 0644); err != nil {
			t.Fatalf("failed to write golden file %s: %v", goldenPath, err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file %s does not exist. Run with -update to create it.", goldenPath)
		}
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Normalize CRLF to LF
	actualClean := bytes.ReplaceAll(actualBytes, []byte("\r\n"), []byte("\n"))
	goldenClean := bytes.ReplaceAll(goldenBytes, []byte("\r\n"), []byte("\n"))

	// If JSON, compare formatted JSON
	if filepath.Ext(goldenPath) == ".json" {
		var actualJSON, goldenJSON any
		if err := json.Unmarshal(actualClean, &actualJSON); err == nil {
			if err := json.Unmarshal(goldenClean, &goldenJSON); err == nil {
				diff := cmp.Diff(goldenJSON, actualJSON)
				if diff != "" {
					t.Errorf("Golden JSON mismatch (-golden +actual):\n%s", diff)
				}
				return
			}
		}
	}

	if !bytes.Equal(goldenClean, actualClean) {
		diff := cmp.Diff(string(goldenClean), string(actualClean))
		t.Errorf("Golden mismatch for %s (-golden +actual):\n%s", goldenPath, diff)
	}
}

// DeleteAll removes the golden directory.
func DeleteAll(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

// VerifyDir compares an entire output directory against a golden directory.
func VerifyDir(t *testing.T, actualDir, goldenDir string) {
	t.Helper()

	err := filepath.Walk(actualDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".pyc") || strings.Contains(path, "__pycache__") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		rel, err := filepath.Rel(actualDir, path)
		if err != nil {
			return err
		}

		goldenFile := filepath.Join(goldenDir, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read actual file %s: %w", path, err)
		}

		t.Run(rel, func(t *testing.T) {
			AssertMatches(t, goldenFile, data)
		})
		return nil
	})

	if err != nil {
		t.Fatalf("failed walking directory %s: %v", actualDir, err)
	}
}
