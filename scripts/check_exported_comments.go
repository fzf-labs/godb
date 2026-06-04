//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	missing, err := missingExportedComments(".")
	if err != nil {
		panic(err)
	}
	for _, item := range missing {
		fmt.Println(item)
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "missing exported comments: %d\n", len(missing))
		os.Exit(1)
	}
}

func missingExportedComments(root string) ([]string, error) {
	fset := token.NewFileSet()
	missing := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if shouldSkip(path, d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !isCheckedGoFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isGenerated(data) {
			return nil
		}
		file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			return err
		}
		missing = append(missing, missingInFile(fset, path, file)...)
		return nil
	})
	sort.Strings(missing)
	return missing, err
}

func shouldSkip(path string, d os.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	base := d.Name()
	return base == ".git" ||
		base == "vendor" ||
		strings.Contains(path, "orm/example/gorm/postgres/gorm_gen_")
}

func isCheckedGoFile(path string) bool {
	return strings.HasSuffix(path, ".go") &&
		!strings.HasSuffix(path, ".gen.go") &&
		!strings.HasSuffix(path, ".pb.go")
}

func isGenerated(data []byte) bool {
	limit := len(data)
	if limit > 512 {
		limit = 512
	}
	return strings.Contains(string(data[:limit]), "Code generated")
}

func missingInFile(fset *token.FileSet, path string, file *ast.File) []string {
	missing := make([]string, 0)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if strings.HasSuffix(path, "_test.go") && isStandardTestEntrypoint(d.Name.Name) {
				continue
			}
			if d.Name.IsExported() && d.Doc == nil {
				missing = append(missing, fmt.Sprintf("%s: exported func %s", fset.Position(d.Pos()), d.Name.Name))
			}
		case *ast.GenDecl:
			missing = append(missing, missingInGenDecl(fset, d)...)
		}
	}
	return missing
}

func isStandardTestEntrypoint(name string) bool {
	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Fuzz") ||
		strings.HasPrefix(name, "Example")
}

func missingInGenDecl(fset *token.FileSet, decl *ast.GenDecl) []string {
	missing := make([]string, 0)
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() && decl.Doc == nil && s.Doc == nil && s.Comment == nil {
				missing = append(missing, fmt.Sprintf("%s: exported type %s", fset.Position(s.Pos()), s.Name.Name))
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() && decl.Doc == nil && s.Doc == nil && s.Comment == nil {
					kind := strings.ToLower(decl.Tok.String())
					missing = append(missing, fmt.Sprintf("%s: exported %s %s", fset.Position(name.Pos()), kind, name.Name))
				}
			}
		}
	}
	return missing
}
