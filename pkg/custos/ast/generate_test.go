package ast

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite zz_generated.go")

// TestGenerated keeps the aliases in step with the syntax package: every
// exported type and constant of its tree, kind and token files.
func TestGenerated(t *testing.T) {
	want := generate(t, filepath.Join("..", "..", "..", "internal", "php", "syntax"))
	if *update {
		if err := os.WriteFile("zz_generated.go", []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile("zz_generated.go")
	if err != nil || string(got) != want {
		t.Fatalf("zz_generated.go is stale: run go test ./pkg/custos/ast -update (%v)", err)
	}
}

func generate(t *testing.T, dir string) string {
	t.Helper()
	var types, consts []string
	fset := token.NewFileSet()
	for _, name := range []string{"ast.go", "kinds.go", "token.go"} {
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, sp := range g.Specs {
				switch s := sp.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() && s.TypeParams == nil {
						types = append(types, s.Name.Name)
					}
				case *ast.ValueSpec:
					if g.Tok == token.CONST {
						for _, n := range s.Names {
							if n.IsExported() {
								consts = append(consts, n.Name)
							}
						}
					}
				}
			}
		}
	}
	sort.Strings(types)
	sort.Strings(consts)
	var b strings.Builder
	b.WriteString("// Code generated from internal/php/syntax; DO NOT EDIT.\n\npackage ast\n\nimport \"github.com/janalis/custos/internal/php/syntax\"\n\n// Types of the syntax tree.\ntype (\n")
	for _, name := range types {
		fmt.Fprintf(&b, "\t%s = syntax.%s\n", name, name)
	}
	b.WriteString(")\n\n// Node kinds, token kinds and the other enumerations of the syntax tree.\nconst (\n")
	for _, name := range consts {
		fmt.Fprintf(&b, "\t%s = syntax.%s\n", name, name)
	}
	b.WriteString(")\n")
	src, err := format.Source([]byte(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}
