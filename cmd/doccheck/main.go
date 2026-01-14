// Package main provides the doccheck CLI entrypoint.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// finding represents a documentation issue found during code scanning.
// It contains the position of the issue in the source file, the type of issue,
// and the name of the undocumented element.
type finding struct {
	pos  token.Position
	kind string
	name string

	file string
}

func main() {
	var root string
	var fail bool
	var includeTests bool

	flag.StringVar(&root, "root", ".", "repository root to scan")
	flag.BoolVar(&fail, "fail", false, "exit non-zero if findings exist")
	flag.BoolVar(&includeTests, "include-tests", false, "include *_test.go files")
	flag.Parse()

	findings, err := scan(root, includeTests)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	for _, f := range findings {
		fmt.Printf("%s:%d:%d: %s: %s\n", relPath(root, f.file), f.pos.Line, f.pos.Column, f.kind, f.name)
	}

	if fail && len(findings) != 0 {
		os.Exit(1)
	}
}

func scan(root string, includeTests bool) ([]finding, error) {
	var findings []finding

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			switch base {
			case ".git", "vendor", "bin", "node_modules", ".next":
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Ignore swagger generated Go sources.
		if strings.Contains(filepath.ToSlash(path), "/docs/swagger/") {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		if !hasPackageComment(file) {
			pos := fset.Position(file.Package)
			findings = append(findings, finding{pos: pos, kind: "missing-package-doc", name: file.Name.Name, file: path})
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(s.Name.Name) && !hasDoc(gen.Doc, s.Doc) {
						pos := fset.Position(s.Pos())
						findings = append(findings, finding{pos: pos, kind: "missing-exported-doc", name: "type " + s.Name.Name, file: path})
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name == nil || !ast.IsExported(name.Name) {
							continue
						}
						if !hasDoc(gen.Doc, s.Doc) {
							pos := fset.Position(name.Pos())
							findings = append(findings, finding{pos: pos, kind: "missing-exported-doc", name: "value " + name.Name, file: path})
							break
						}
					}
				}
			}
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv != nil {
				continue // methods are optional for now
			}
			if fn.Name == nil || !ast.IsExported(fn.Name.Name) {
				continue
			}
			if fn.Doc == nil || len(fn.Doc.List) == 0 {
				pos := fset.Position(fn.Pos())
				findings = append(findings, finding{pos: pos, kind: "missing-exported-doc", name: "func " + fn.Name.Name, file: path})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func hasPackageComment(file *ast.File) bool {
	if file == nil {
		return false
	}
	for _, cg := range file.Comments {
		if cg == nil {
			continue
		}
		if cg.End() < file.Package {
			for _, c := range cg.List {
				if c == nil {
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(c.Text), "// Package ") {
					return true
				}
			}
		}
	}
	return false
}

func hasDoc(groups ...*ast.CommentGroup) bool {
	for _, g := range groups {
		if g == nil {
			continue
		}
		if len(g.List) != 0 {
			return true
		}
	}
	return false
}

func relPath(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}
