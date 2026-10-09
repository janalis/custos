package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, path, src string) {
	t.Helper()
	path = filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func layout(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for path, src := range map[string]string{
		"internal/inspection/meta/rules.json": `[{"id":"Example"}]`,
		"internal/inspection/rules/example/rule.go": `package example
import "custos/internal/inspection/analysis"
type rule struct{}
func (rule) ID() string { return "Example" }
func New() analysis.Rule { return rule{} }
`,
		"internal/inspection/catalogue/catalogue.go": `package catalogue
import inspection "custos/internal/inspection/rules/example"
func All() []any { return []any{inspection.New()} }
`,
		"cmd/custos/main.go":               "package main\nimport \"custos/internal/cli\"\nfunc main(){cli.Run()}",
		"specs/Example.md":                 "# Example\n",
		"testdata/rules/Example/basic.php": "<?php\n",
		"internal/ignored_test.go":         "not Go; test files are excluded",
		"internal/asset.txt":               "asset",
	} {
		write(t, root, path, src)
	}
	return root
}

func TestCheck(t *testing.T) {
	root := layout(t)
	problems, err := check(root)
	if err != nil || len(problems) != 0 {
		t.Fatalf("%v: %v", problems, err)
	}
}

func TestViolations(t *testing.T) {
	cases := []struct{ name, path, src, want string }{
		{"dependency", "internal/semantic/types/bad.go", "package types\nimport \"custos/internal/output/report\"", "forbidden dependency"},
		{"cross rule", "internal/inspection/rules/example/bad.go", "package example\nimport \"custos/internal/inspection/rules/other\"", "forbidden dependency"},
		{"init", "internal/inspection/rules/example/init.go", "package example\nfunc init(){}", "cannot register"},
		{"duplicate ID", "internal/inspection/rules/other/rule.go", "package other\ntype rule struct{}\nfunc (rule) ID()string{return \"Example\"}", "duplicate inspection ID"},
		{"unknown ID", "internal/inspection/rules/other/rule.go", "package other\ntype rule struct{}\nfunc (rule) ID()string{return \"Other\"}", "unknown inspection ID"},
		{"wrong directory", "internal/inspection/rules/wrong/rule.go", "package wrong\ntype rule struct{}\nfunc (rule) ID()string{return \"Odd\"}", "directory must be lowercase"},
		{"missing constructor", "internal/inspection/rules/example/rule.go", "package example\ntype rule struct{}\nfunc (rule) ID()string{return \"Example\"}", "New constructor missing"},
		{"missing registration", "internal/inspection/catalogue/catalogue.go", "package catalogue\nfunc All() []any{return nil}", "exactly once (got 0)"},
		{"duplicate registration", "internal/inspection/catalogue/catalogue.go", "package catalogue\nimport \"custos/internal/inspection/rules/example\"\nfunc All() []any{return []any{example.New(),example.New()}}", "exactly once (got 2)"},
		{"missing implementation", "internal/inspection/meta/rules.json", `[{"id":"Missing"}]`, "implementation missing"},
		{"empty ID", "internal/inspection/rules/example/rule.go", "package example\ntype rule struct{}\nfunc (rule) ID()string{ return \"\" }\nfunc New()any{return rule{}}", "unknown inspection ID"},
		{"missing fixture", "testdata/rules/Example/basic.php", "fixture", ""},
		{"opaque ID", "internal/inspection/rules/example/rule.go", "package example\ntype rule struct{}\nfunc (rule) ID()string{ if true {return}; return name }\nfunc (rule) OtherID()int{return 1}", "implementation missing"},
		{"numeric ID", "internal/inspection/rules/example/rule.go", "package example\ntype rule struct{}\nfunc (rule) ID()int{return 123}", "implementation missing"},
		{"catalogue expression", "internal/inspection/catalogue/catalogue.go", "package catalogue\nimport \"custos/internal/inspection/rules/example\"\nfunc All() []any { f(); example.Other(); factory.New().New(); return nil }", "exactly once"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := layout(t)
			write(t, root, tc.path, tc.src)
			problems, err := check(root)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != "" && !strings.Contains(strings.Join(problems, "\n"), tc.want) {
				t.Fatalf("%v", problems)
			}
		})
	}
}

func TestMissingArtifacts(t *testing.T) {
	for _, path := range []string{"specs/Example.md", "testdata/rules/Example"} {
		t.Run(path, func(t *testing.T) {
			root := layout(t)
			if err := os.RemoveAll(filepath.Join(root, path)); err != nil {
				t.Fatal(err)
			}
			problems, err := check(root)
			if err != nil || len(problems) == 0 {
				t.Fatalf("%v: %v", problems, err)
			}
		})
	}
	root := layout(t)
	if err := os.Remove(filepath.Join(root, "specs/Example.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "specs/Example.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "testdata/rules/Example/basic.php")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "testdata/rules/Example/readme.txt", "empty")
	if err := os.Mkdir(filepath.Join(root, "testdata/rules/Example/subdir.php"), 0o755); err != nil {
		t.Fatal(err)
	}
	problems, err := check(root)
	if err != nil || len(problems) != 2 {
		t.Fatalf("%v: %v", problems, err)
	}
}

func TestRun(t *testing.T) {
	var out, stderr bytes.Buffer
	if got := run([]string{"-bad"}, &out, &stderr); got != 2 {
		t.Fatal(got)
	}
	if got := run([]string{"-root", t.TempDir()}, &out, &stderr); got != 2 {
		t.Fatal(got)
	}
	root := layout(t)
	if got := run([]string{"-root", root}, &out, &stderr); got != 0 {
		t.Fatalf("%d: %s", got, stderr.String())
	}
	write(t, root, "internal/php/syntax/bad.go", "package syntax\nimport \"custos/internal/cli\"")
	if got := run([]string{"-root", root}, &out, &stderr); got != 1 {
		t.Fatal(got)
	}
}

func TestErrors(t *testing.T) {
	for _, path := range []string{"internal/inspection/meta/rules.json", "internal/bad.go"} {
		root := layout(t)
		write(t, root, path, "invalid")
		if _, err := check(root); err == nil {
			t.Fatal("expected error")
		}
	}
	root := layout(t)
	if err := os.RemoveAll(filepath.Join(root, "cmd")); err != nil {
		t.Fatal(err)
	}
	if _, err := check(root); err == nil {
		t.Fatal("expected walk error")
	}
}

func TestBoundaries(t *testing.T) {
	for _, from := range []string{"internal/php/syntax", "internal/semantic/infer", "internal/diagnostic", "internal/fixing", "internal/inspection/meta", "internal/inspection/analysis", "internal/inspection/astquery", "internal/inspection/flowquery", "internal/inspection/rules/example", "internal/inspection/catalogue", "internal/inspection/semanticquery", "internal/inspection/phpunit", "internal/project", "internal/project/baseline", "internal/output/report", "internal/platform/safeio"} {
		if allowed(from, "internal/cli") {
			t.Fatalf("%s permits CLI", from)
		}
	}
	for _, pair := range [][2]string{{"internal/php/syntax", "internal/php/version"}, {"internal/semantic/infer", "internal/semantic/types"}, {"internal/diagnostic", "internal/php/syntax"}, {"internal/fixing", "internal/diagnostic"}, {"internal/inspection/meta", "internal/diagnostic"}, {"internal/inspection/analysis", "internal/semantic/infer"}, {"internal/inspection/astquery", "internal/php/syntax"}, {"internal/inspection/astquery", "internal/diagnostic"}, {"internal/inspection/flowquery", "internal/inspection/astquery"}, {"internal/inspection/rules/example", "internal/inspection/phpunit"}, {"internal/inspection/catalogue", "internal/inspection/rules/example"}, {"internal/inspection/semanticquery", "internal/inspection/flowquery"}, {"internal/project", "internal/fixing"}, {"internal/output/report", "internal/diagnostic"}, {"internal/platform/safeio", "internal/platform/safeio"}, {"cmd/custos", "internal/cli"}, {"internal/editor/lsp", "internal/project"}} {
		if !allowed(pair[0], pair[1]) {
			t.Fatal(pair)
		}
	}
}

func TestConstructor(t *testing.T) {
	for _, src := range []string{"func New() analysis.Rule{return nil}", "func Other() analysis.Rule{return nil}", "func (r rule) New() analysis.Rule{return nil}", "func New(arg int) analysis.Rule{return nil}", "func New(){}", "func New()(analysis.Rule,error){return nil,nil}", "func New() int{return 0}", "func New() analysis.Other{return nil}", "func New() fake.Rule{return nil}"} {
		file, err := parser.ParseFile(token.NewFileSet(), "test.go", "package example\n"+src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := constructor(file.Decls[0].(*ast.FuncDecl), map[string]string{"analysis": "custos/internal/inspection/analysis"})
		if got != (src == "func New() analysis.Rule{return nil}") {
			t.Fatalf("%s: %v", src, got)
		}
	}
}

func TestOrphanRuleDirectory(t *testing.T) {
	root := layout(t)
	write(t, root, "internal/inspection/rules/orphan/helper.go", "package orphan\nconst unused = 1")
	problems, err := check(root)
	if err != nil || !strings.Contains(strings.Join(problems, "\n"), "rule directory has no inspection ID") {
		t.Fatalf("%v: %v", problems, err)
	}
}
