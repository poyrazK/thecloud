package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsMissingDocs(t *testing.T) {
	root := t.TempDir()

	badFile := filepath.Join(root, "bad.go")
	badSource := `package bad

type ExportedType struct{}

// ExportedFunc has docs.
func ExportedFunc() {}
`
	if err := os.WriteFile(badFile, []byte(badSource), 0600); err != nil {
		t.Fatalf("write bad.go: %v", err)
	}

	badTypeFile := filepath.Join(root, "badtype.go")
	badTypeSource := `// Package badtype demonstrates missing docs.
package badtype

type MissingDoc struct{}
`
	if err := os.WriteFile(badTypeFile, []byte(badTypeSource), 0600); err != nil {
		t.Fatalf("write badtype.go: %v", err)
	}

	badTestFile := filepath.Join(root, "ignored_test.go")
	badTestSource := `package ignored

type IgnoredType struct{}
`
	if err := os.WriteFile(badTestFile, []byte(badTestSource), 0600); err != nil {
		t.Fatalf("write ignored_test.go: %v", err)
	}

	findings, err := scan(root, false)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(findings) == 0 {
		t.Fatalf("expected findings, got none")
	}

	for _, f := range findings {
		if filepath.Base(f.file) == filepath.Base(badTestFile) {
			t.Fatalf("did not expect findings from test files when includeTests=false")
		}
	}
}

func TestScanIncludesTestsWhenEnabled(t *testing.T) {
	root := t.TempDir()

	badTestFile := filepath.Join(root, "included_test.go")
	badTestSource := `package included

type MissingDoc struct{}
`
	if err := os.WriteFile(badTestFile, []byte(badTestSource), 0600); err != nil {
		t.Fatalf("write included_test.go: %v", err)
	}

	findings, err := scan(root, true)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(findings) == 0 {
		t.Fatalf("expected findings, got none")
	}

	foundTest := false
	for _, f := range findings {
		if filepath.Base(f.file) == filepath.Base(badTestFile) {
			foundTest = true
			break
		}
	}
	if !foundTest {
		t.Fatalf("expected findings from test files when includeTests=true")
	}
}

func TestRelPathFallback(t *testing.T) {
	path := relPath("", "/tmp/file.go")
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}
