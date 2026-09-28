package unit

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skipDirs are directories (given as paths relative to the repo root)
// TestStrayRefs does not walk into: historical records (spec artifacts, past
// review reports) that are not rewritten, and tooling/VCS state that isn't
// source.
var skipDirs = map[string]bool{
	".git":         true,
	"specs":        true,
	"docs/reviews": true,
	".specify":     true,
	".idea":        true,
	".claude":      true,
	"trash":        true,
}

// scannedExts are file extensions TestStrayRefs scans wholesale.
var scannedExts = map[string]bool{".go": true, ".sh": true, ".yaml": true, ".yml": true, ".md": true}

// scannedBasenames are extension-less or otherwise-named files TestStrayRefs
// scans by exact basename.
var scannedBasenames = map[string]bool{"Makefile": true}

func shouldScan(path string) bool {
	base := filepath.Base(path)
	if scannedBasenames[base] {
		return true
	}
	return scannedExts[filepath.Ext(base)]
}

// TestStrayRefs guards against leftover references to the pre-rename module
// path or the pre-move hub package's old location under internal/ surviving
// anywhere in the live tree. The search patterns are built by concatenation,
// and this comment avoids spelling them out contiguously, so this file's own
// text doesn't trip the check.
func TestStrayRefs(t *testing.T) {
	oldModule := `"` + "sonora-cli" + "/"
	oldLdflags := "X " + "sonora-cli" + "/"
	oldHubPkg := "internal/" + "hub"

	root := "../.."
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == root {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if skipDirs[filepath.ToSlash(rel)] {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldScan(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if strings.Contains(line, oldModule) || strings.Contains(line, oldLdflags) || strings.Contains(line, oldHubPkg) {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s:%d: stray reference to a pre-refactor path: %s", filepath.ToSlash(rel), lineNo, strings.TrimSpace(line))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
}
