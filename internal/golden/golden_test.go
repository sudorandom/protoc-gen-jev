package golden_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sudorandom/protoc-gen-jev/internal/golden"
)

func TestGoldenOutputs(t *testing.T) {
	// Root project directories
	actualJevDir := filepath.Join("..", "..", "gen", "jev")
	goldenJevDir := filepath.Join("..", "..", "testdata", "golden")

	if _, err := os.Stat(actualJevDir); os.IsNotExist(err) {
		t.Fatalf("generated directory %s does not exist. Run buf generate first.", actualJevDir)
	}

	golden.VerifyDir(t, actualJevDir, goldenJevDir)
}
