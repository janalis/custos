package filepermissionfailurepassespolicy

import (
	"os"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\nif ((fileperms($path) & 0002) === 0) { return true; }\n", 1},
		{"<?php\n$m = fileperms($path); if ($m !== false && ($m & 0002) === 0) { return true; }\n", 0},
		{"<?php other();", 0},
		{"<?php if((fileperms($p)&$mask)===0){return true;}", 0},
		{"<?php if((fileperms($p)&0)===0){return true;}", 0},
		{"<?php if(($mode&2)===0){return true;}", 0},
		{"<?php if((fileperms($p)&2)===0){throw new Exception;}", 0},
		{"<?php if((fileperms($p)&2)===0){echo 'ok';}", 0},
		{"<?php $ok=(fileperms($p)&2)===0;", 0},
		{"<?php if((fileperms($p)&2)==0){return true;}", 1},
		{"<?php if((fileperms($p)&2)===1){return true;}", 0},
		{"<?php function unreachable(){return;\nif ((fileperms($path) & 0002) === 0) { return true; }\n}", 0},
		{"<?php \nif ((fileperms($path) & 0002) === 0) { return true; }\n function broken(", 0},
		{"<?php $r=fileperms($p);$alias=&$r;mutate($alias);if(($r&2)===0){return true;}", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}

func TestFixtureRanges(t *testing.T) {
	for _, name := range []string{"basic", "negative", "guards", "false-positives"} {
		marked, err := os.ReadFile("testdata/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", src, syntax.Options{}))
		if len(got) != len(want) {
			t.Fatalf("%s: got %+v, want %+v", name, got, want)
		}
		for i, g := range got {
			w := want[i]
			if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity || len(g.Fixes) != 0 {
				t.Fatalf("got %+v, want %+v", g, w)
			}
		}
	}
}

func TestReviewGuards(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
	if err != nil {
		t.Fatal(err)
	}
	const use = " if (($p & 0222) === 0) { return true; } return false; }"
	for _, tc := range []struct {
		name, guard string
		want        int
	}{
		{"none", "", 1},
		{"falsy", "if (!$p) { return false; }", 0},
		{"parenthesised", "if ((!($p))) { return false; }", 0},
		{"empty", "if (empty($p)) { return false; }", 0},
		{"not int", "if (!is_int($p)) { return false; }", 0},
		{"not integer alias", "if (!\\is_integer($p)) { throw new Exception; }", 0},
		{"bool", "if (is_bool($p)) { return false; }", 0},
		{"loose", "if ($p == false) { return false; }", 0},
		{"strict yoda", "if (false === $p) { return false; }", 0},
		{"earlier sibling", "if (!$p) { return false; } echo $f;", 0},
		{"non-terminating", "if (!$p) { echo $f; }", 1},
		{"inverted", "if ($p) { return false; }", 1},
		{"inverted type", "if (is_int($p)) { return false; }", 1},
		{"other variable", "if (!$f) { return false; }", 1},
		{"reassigned", "if (!$p) { return false; } $p = fileperms($f . 'x');", 1},
		{"conjunction", "if (!$p && $f) { return false; }", 1},
		{"named argument", "if (!is_int(value: $p)) { return false; }", 0},
		{"null comparison", "if ($p === null) { return false; }", 1},
	} {
		src := "<?php function check($f) { $p = fileperms($f); " + tc.guard + use
		f := syntax.Parse("test.php", []byte(src), syntax.Options{})
		if len(f.Errors) != 0 {
			t.Fatalf("%s: invalid probe: %v", tc.name, f.Errors)
		}
		if got := e.Analyze(f); len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.name, len(got), tc.want, got)
		}
	}
}

// A yield in the scope leaves the flow facts incomplete, so only the guard
// walk can relate a loose or truthiness test to the later mask check.
func TestGuardsWithoutFlowFacts(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilePermissionFailurePassesPolicy"}})
	if err != nil {
		t.Fatal(err)
	}
	const check = "if (($p & 0222) === 0) { return true; }"
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"none", check, 1},
		{"and operand", "if ($p && ($p & 0222) === 0) { return true; }", 0},
		{"branch", "if ($p) { " + check + " }", 0},
		{"elseif", "if ($f === '') { return false; } elseif ($p) { " + check + " }", 0},
		{"else", "if (!$p) { echo $f; } else { " + check + " }", 0},
		{"else after elseif", "if ($f === '') { echo 1; } elseif (!$p) { echo 2; } else { " + check + " }", 0},
		{"elseif after failed test", "if ($f === '') { echo 1; } elseif (!$p) { echo 2; } elseif ($f) { " + check + " }", 0},
		{"earlier exit", "if (!$p) { return false; } " + check, 0},
		{"loose exit", "if ($p == false) { return false; } " + check, 0},
		{"invalid predicate arity", "if (is_int()) { " + check + " }", 1},
		{"disjunction exit", "if (!$p || $f === '') { return false; } " + check, 0},
		{"wrong branch", "if (!$p) { " + check + " }", 1},
		{"unrelated else", "if ($f === '') { echo 1; } elseif ($f) { echo 2; } else { " + check + " }", 1},
	} {
		src := "<?php function check($f) { $p = fileperms($f); yield 0; " + tc.body + " return false; }"
		f := syntax.Parse("test.php", []byte(src), syntax.Options{})
		if len(f.Errors) != 0 {
			t.Fatalf("%s: invalid probe: %v", tc.name, f.Errors)
		}
		if got := e.Analyze(f); len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.name, len(got), tc.want, got)
		}
	}
}
