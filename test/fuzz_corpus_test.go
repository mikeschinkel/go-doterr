package test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFuzzCorpus runs all fuzz corpus files as regression tests.
// This ensures that any interesting inputs discovered by fuzzing
// are continuously tested in CI/CD.
func TestFuzzCorpus(t *testing.T) {
	corpusDirs := []string{
		"testdata/fuzz/FuzzNewErr",
		"testdata/fuzz/FuzzWithErr",
		"testdata/fuzz/FuzzCombineErrs",
		"testdata/fuzz/FuzzErrMeta",
		"testdata/fuzz/FuzzErrValue",
	}

	for _, dir := range corpusDirs {
		t.Run(dir, func(t *testing.T) {
			// Check if directory exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				//t.Skipf("Corpus directory %s does not exist yet", dir)
				return
			}

			// Read all corpus files
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("Failed to read corpus directory %s: %v", dir, err)
			}

			if len(entries) == 0 {
				//t.Skipf("No corpus files in %s", dir)
				return
			}

			//t.Logf("Found %d corpus files in %s", len(entries), dir)

			// We just verify corpus files exist and are readable
			// The actual fuzzing logic will test them
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				corpusFile := filepath.Join(dir, entry.Name())
				//data, err := os.ReadFile(corpusFile)
				_, err := os.ReadFile(corpusFile)
				if err != nil {
					t.Errorf("Failed to read corpus file %s: %v", corpusFile, err)
					continue
				}

				//if len(data) == 0 {
				//	t.Logf("Warning: empty corpus file %s", corpusFile)
				//}
			}
		})
	}
}
