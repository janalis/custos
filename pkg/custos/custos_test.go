package custos

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/pkg/custos/ast"
)

func noEval(t *testing.T) Rule {
	t.Helper()
	return MustRule(Info{ID: "NoEval", Severity: SeverityError, Summary: "eval runs arbitrary code.", HasFix: true, PHPMin: "7.0"},
		[]ast.NodeKind{ast.KEval},
		func(ctx *Context, n ast.Node) {
			e := n.(*ast.Eval)
			ctx.ReportNode(n, "Remove eval of "+ctx.Text(e.Expr), Fix{Title: "Remove", Edits: func() []TextEdit {
				return []TextEdit{{Span: n.Span(), NewText: "null"}}
			}})
		})
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.php"), []byte("<?php\n$a = eval('1;');\ndump(2);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	dump, err := Pattern(PatternDefinition{ID: "NoDump", Match: Patterns{`\dump(...)`}, Message: "Remove the dump"})
	if err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := Run("test", []string{"analyse", "--no-cache", "--rule", "NoEval,NoDump", "."}, strings.NewReader(""), &out, &errb, noEval(t), dump)
	if code != 1 || !strings.Contains(out.String(), "error: Remove eval of '1;' [NoEval] (fixable)") || !strings.Contains(out.String(), "[NoDump]") {
		t.Fatalf("analyse: %d %s %s", code, out.String(), errb.String())
	}
	out.Reset()
	if code := Run("test", []string{"explain", "NoEval"}, strings.NewReader(""), &out, &errb, noEval(t)); code != 0 || !strings.Contains(out.String(), "group: Custom, severity: error, enabled by default, has a quick-fix") {
		t.Fatalf("explain: %d %s", code, out.String())
	}
	if code := Run("test", []string{"fix", "--no-cache", "--rule", "NoEval"}, strings.NewReader(""), &out, &errb, noEval(t)); code != 0 {
		t.Fatalf("fix: %d %s", code, errb.String())
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "a.php")); string(got) != "<?php\n$a = null;\ndump(2);\n" {
		t.Fatalf("fixed: %q", got)
	}
	// A custom build is not a custos release: it does not replace itself.
	errb.Reset()
	if code := Run("1.0.0", []string{"self-update"}, strings.NewReader(""), &out, &errb, noEval(t)); code != 2 || !strings.Contains(errb.String(), "rebuild this program") {
		t.Fatalf("self-update: %d %s", code, errb.String())
	}
	// The extra rules last for one run only.
	out.Reset()
	if code := Run("test", []string{"explain", "NoEval"}, strings.NewReader(""), &out, &errb); code != 2 {
		t.Fatalf("explain without the rule: %d", code)
	}
}

func TestMain_(t *testing.T) {
	defer func() { exit = os.Exit }()
	got := -1
	exit = func(code int) { got = code }
	args := os.Args
	defer func() { os.Args = args }()
	os.Args = []string{"custos", "version"}
	Main("test", noEval(t))
	if got != 0 {
		t.Fatalf("exit %d", got)
	}
}

func TestNewRule(t *testing.T) {
	check := func(*Context, ast.Node) {}
	kinds := []ast.NodeKind{ast.KEval}
	for _, tc := range []struct {
		info  Info
		kinds []ast.NodeKind
		check func(*Context, ast.Node)
		want  string
	}{
		{Info{ID: "lower"}, kinds, check, "must start with an uppercase letter"},
		{Info{ID: "StaticClosureCanBeUsed"}, kinds, check, "names a built-in rule"},
		{Info{ID: "R", Severity: "fatal"}, kinds, check, "invalid severity"},
		{Info{ID: "R", PHPMin: "9"}, kinds, check, "invalid version"},
		{Info{ID: "R"}, nil, check, "needs node kinds"},
		{Info{ID: "R"}, kinds, nil, "needs node kinds"},
	} {
		if _, err := NewRule(tc.info, tc.kinds, tc.check); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%+v: %v", tc.info, err)
		}
	}
	r, err := NewRule(Info{ID: "R", Disabled: true}, kinds, check)
	if err != nil {
		t.Fatal(err)
	}
	f := r.(Described).Facts()
	if f.Severity != SeverityWarning || f.EnabledByDefault || f.Group != Group || r.ID() != "R" || len(r.Kinds()) != 1 {
		t.Fatalf("facts %+v", f)
	}
	if d := r.(Described).Description(); d.Fix || d.Summary != "" {
		t.Fatalf("description %+v", d)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustRule did not panic")
		}
	}()
	MustRule(Info{ID: "bad"}, kinds, check)
}
