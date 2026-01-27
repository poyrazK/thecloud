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

const (
	missingPackageDocKind  = "missing-package-doc"
	missingExportedDocKind = "missing-exported-doc"
)

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
	findings := make([]finding, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipFile(path, includeTests) {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		appendMissingPackageDoc(&findings, fset, file, path)
		appendMissingExportedDeclDocs(&findings, fset, file, path)
		appendMissingExportedFuncDocs(&findings, fset, file, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func shouldSkipDir(path string) bool {
	base := filepath.Base(path)
	switch base {
	case ".git", "vendor", "bin", "node_modules", ".next":
		return true
	default:
		return false
	}
}

func shouldSkipFile(path string, includeTests bool) bool {
	if !strings.HasSuffix(path, ".go") {
		return true
	}
	if !includeTests && strings.HasSuffix(path, "_test.go") {
		return true
	}
	return strings.Contains(filepath.ToSlash(path), "/docs/swagger/")
}

func appendMissingPackageDoc(findings *[]finding, fset *token.FileSet, file *ast.File, path string) {
	if hasPackageComment(file) {
		return
	}
	pos := fset.Position(file.Package)
	*findings = append(*findings, finding{pos: pos, kind: missingPackageDocKind, name: file.Name.Name, file: path})
}

func appendMissingExportedDeclDocs(findings *[]finding, fset *token.FileSet, file *ast.File, path string) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				appendMissingTypeDoc(findings, fset, gen, s, path)
			case *ast.ValueSpec:
				appendMissingValueDoc(findings, fset, gen, s, path)
			}
		}
	}
}

func appendMissingTypeDoc(findings *[]finding, fset *token.FileSet, gen *ast.GenDecl, spec *ast.TypeSpec, path string) {
	if !ast.IsExported(spec.Name.Name) || hasDoc(gen.Doc, spec.Doc) {
		return
	}
	pos := fset.Position(spec.Pos())
	*findings = append(*findings, finding{pos: pos, kind: missingExportedDocKind, name: "type " + spec.Name.Name, file: path})
}

func appendMissingValueDoc(findings *[]finding, fset *token.FileSet, gen *ast.GenDecl, spec *ast.ValueSpec, path string) {
	if hasDoc(gen.Doc, spec.Doc) {
		return
	}
	name := firstExportedName(spec.Names)
	if name == nil {
		return
	}
	pos := fset.Position(name.Pos())
	*findings = append(*findings, finding{pos: pos, kind: missingExportedDocKind, name: "value " + name.Name, file: path})
}

func firstExportedName(names []*ast.Ident) *ast.Ident {
	for _, name := range names {
		if name != nil && ast.IsExported(name.Name) {
			return name
		}
	}
	return nil
}

func appendMissingExportedFuncDocs(findings *[]finding, fset *token.FileSet, file *ast.File, path string) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Recv != nil {
			continue
		}
		if fn.Name == nil || !ast.IsExported(fn.Name.Name) {
			continue
		}
		if fn.Doc == nil || len(fn.Doc.List) == 0 {
			pos := fset.Position(fn.Pos())
			*findings = append(*findings, finding{pos: pos, kind: missingExportedDocKind, name: "func " + fn.Name.Name, file: path})
		}
	}
}

func hasPackageComment(file *ast.File) bool {
	if file == nil {
		return false
	}
	for _, cg := range file.Comments {
		if isPackageCommentGroup(file, cg) {
			return true
		}
	}
	return false
}

func isPackageCommentGroup(file *ast.File, group *ast.CommentGroup) bool {
	if group == nil || group.End() >= file.Package {
		return false
	}
	for _, c := range group.List {
		if c == nil {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(c.Text), "// Package ") {
			return true
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
