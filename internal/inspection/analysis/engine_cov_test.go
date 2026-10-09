package analysis

import (
	"reflect"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// optRule reads every option accessor once per file and records the values.
type optRule struct{ got *map[string]any }

func (optRule) ID() string               { return "NestedNotOperators" }
func (optRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (r optRule) Check(ctx *Context, n syntax.Node) {
	m := map[string]any{
		"rule": ctx.RuleID(),
		"b":    ctx.Bool("b"), "bs": ctx.Bool("bs"), "bn": ctx.Bool("n"),
		"i": ctx.Int("i"), "if": ctx.Int("f"), "is": ctx.Int("is"), "in": ctx.Int("b"),
		"s": ctx.String("str"), "sn": ctx.String("i"),
		"l": ctx.List("l"), "la": ctx.List("la"), "ln": ctx.List("i"),
		"set": ctx.OptionSet("i"), "unset": ctx.OptionSet("zz"),
		"text":  ctx.Text(n),
		"nil":   ctx.Text(nil),
		"memo1": ctx.Memo("k", func() any { return 1 }),
		"memo2": ctx.Memo("k", func() any { return 2 }),
	}
	*r.got = m
	ctx.ReportNode(n, "x")
}

func TestOptionAccessorsAndSeverity(t *testing.T) {
	var got map[string]any
	cfg := Config{Rules: map[string]RuleConfig{"NestedNotOperators": {
		Severity: diagnostic.SeverityError,
		Options: map[string]any{
			"b": true, "bs": "true", "i": 3, "f": float64(2), "is": "7",
			"str": "v", "l": []string{"a"}, "la": []any{"x", 1},
		},
	}}}
	e, err := NewEngine([]Rule{optRule{&got}}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(e.Rules(), []string{"NestedNotOperators"}) || e.Config().Rules == nil {
		t.Fatalf("rules %v", e.Rules())
	}
	fs := e.Analyze(syntax.Parse("x.php", []byte("<?php $a = !$b;"), syntax.Options{Version: phpversion.PHP84}))
	if len(fs) != 1 || fs[0].Severity != diagnostic.SeverityError {
		t.Fatalf("severity override not applied: %+v", fs)
	}
	want := map[string]any{
		"rule": "NestedNotOperators", "b": true, "bs": true, "bn": false,
		"i": 3, "if": 2, "is": 7, "in": 0, "s": "v", "sn": "",
		"l": []string{"a"}, "la": []string{"x", "1"}, "ln": []string(nil),
		"set": true, "unset": false, "text": "!$b", "nil": "", "memo1": 1, "memo2": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
}

func TestEngineSelection(t *testing.T) {
	if _, err := NewEngine([]Rule{patternRule{"NoSuchRule"}}, Config{}); err == nil {
		t.Fatal("unknown rule must be refused")
	}
	e, _ := NewEngine([]Rule{okRule{}, patternRule{"SecurityAdvisories"}}, Config{Only: []string{"NestedNotOperators"}})
	if !reflect.DeepEqual(e.Rules(), []string{"NestedNotOperators"}) || e.NeedsIndex() {
		t.Fatalf("only: %v", e.Rules())
	}
	if e.FilePatterns() != nil {
		t.Fatal("no pattern rule enabled")
	}
	e, _ = NewEngine([]Rule{okRule{}, semRule{}}, Config{EnableAll: true})
	if !e.NeedsIndex() {
		t.Fatal("semantic rule needs the index")
	}
	e = e.WithIndex(nil)
	if e.projectIndex() == nil {
		t.Fatal("stubs fallback")
	}
	if c := NewTestContext(e, syntax.Parse("x.php", []byte("<?php"), syntax.Options{})); c.RuleID() != "NestedNotOperators" {
		t.Fatal("test context runs the first rule")
	}
}

type semRule struct{ okRule }

func (semRule) ID() string { return "UnnecessarySemicolon" }
func (semRule) Semantic()  {}

// TestFileRuleDisabledAfterCrash: a crashing file rule is skipped on the
// retry while node rules keep running.
type crashFileRule struct{ patternRule }

func (crashFileRule) CheckFile(*Context) { panic("boom") }

func TestFileRuleDisabledAfterCrash(t *testing.T) {
	e, _ := NewEngine([]Rule{crashFileRule{patternRule{"SecurityAdvisories"}}, okRule{}}, Config{EnableAll: true})
	got := e.Analyze(syntax.Parse("x.php", []byte("<?php $a = !$b;"), syntax.Options{Version: phpversion.PHP84}))
	if len(got) != 2 || got[0].Rule != "internal" || got[1].Rule != "NestedNotOperators" {
		t.Fatalf("got %+v", got)
	}
}

// A panic outside any rule (here: a nil statement in a hand-built tree) is
// not swallowed as a rule failure.
func TestNonRulePanicPropagates(t *testing.T) {
	e, _ := NewEngine([]Rule{okRule{}}, Config{EnableAll: true})
	defer func() {
		if recover() == nil {
			t.Fatal("expected the panic to propagate")
		}
	}()
	e.Analyze(&syntax.File{Stmts: []syntax.Stmt{nil}})
}

func TestSortFindingsTieBreak(t *testing.T) {
	fs := []diagnostic.Finding{
		{Rule: "B", Span: syntax.Span{Start: 1, End: 5}},
		{Rule: "A", Span: syntax.Span{Start: 1, End: 5}},
		{Rule: "C", Span: syntax.Span{Start: 1, End: 3}},
	}
	sortFindings(fs)
	if fs[0].Rule != "C" || fs[1].Rule != "A" || fs[2].Rule != "B" {
		t.Fatalf("got %+v", fs)
	}
}

func TestFunctionName(t *testing.T) {
	e, _ := NewEngine(nil, Config{})
	f := syntax.Parse("x.php", []byte("<?php namespace N; Strlen($x); \\Foo\\Bar($x); $f($x);"), syntax.Options{Version: phpversion.PHP84})
	ctx := NewTestContext(e, f)
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			got = append(got, ctx.FunctionName(call))
		}
		return true
	})
	if !reflect.DeepEqual(got, []string{"strlen", "foo\\bar", ""}) {
		t.Fatalf("got %q", got)
	}
}
