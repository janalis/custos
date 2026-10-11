package arraysplicediscardsreplacementkeys

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArraySpliceDiscardsReplacementKeys" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"strlen(\"x\");", 0}, {"array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[]; array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[]; echo array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[\"k\"=>1]; array_splice($a,0,0,[\"k\"=>2]); echo $a[\"k\"];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $a[\"k\"];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $a[$unknown];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $b[\"q\"];", 0}, {"array_splice($obj->p,0,0,[\"k\"=>1]);", 0}, {"$a=[];array_splice($a,0,0,[\"k\"=>1]);echo $a[\"k\"],1;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArraySpliceDiscardsReplacementKeys"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}

// A function local that only plain assignments can change keeps its tracked
// content across unrelated statements and branches; any other use of the
// variable, and any scope where it is not private, withdraws the proof.
func TestPrivateLocalProof(t *testing.T) {
	const splice = "array_splice($a,0,0,[\"name\"=>\"Ada\"]); echo $a[\"name\"];"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"unrelated statement", "function f(){ $a=[]; $b=[]; " + splice + " }", 1},
		{"unrelated call", "function f(){ $a=[]; g(); " + splice + " }", 1},
		{"branch", "function f($c){ $a=[]; if($c){ " + splice + " } }", 1},
		{"method", "class C { function f($c){ $a=[]; if($c){ " + splice + " } } }", 1},
		{"closure", "$f=function($c){ $a=[]; if($c){ " + splice + " } };", 1},
		{"reassigned with key", "function f($c){ $a=[]; $a=[\"name\"=>1]; g(); " + splice + " }", 0},
		{"maybe reassigned with key", "function f($c){ $a=[]; if($c){ $a=[\"name\"=>1]; } " + splice + " }", 0},
		{"loop reassigns with key", "function f($xs){ $a=[]; foreach($xs as $x){ " + splice + " $a=[\"name\"=>1]; } }", 0},
		{"top-level call", "$a=[]; g(); " + splice, 0},
		{"property hook", "class C { public $p { get { $a=[]; g(); " + splice + " } } }", 0},
		{"parameter", "function f(array &$a){ $a=[]; g(); " + splice + " }", 0},
		{"captured in scope closure", "$f=function() use (&$a){ $a=[]; g(); " + splice + " };", 0},
		{"other capture", "$f=function() use (&$b){ $a=[]; g(); " + splice + " };", 1},
		{"captured by nested closure", "function f(){ $a=[]; $s=function() use (&$a){ $a[\"name\"]=1; }; $s(); " + splice + " }", 0},
		{"global", "function f(){ global $a; $a=[]; g(); " + splice + " }", 0},
		{"static", "function f(){ static $a; $a=[]; g(); " + splice + " }", 0},
		{"alias", "function f(){ $a=[]; $b=&$a; $b[\"name\"]=1; " + splice + " }", 0},
		{"reference assignment", "function f(array &$o){ $a=[]; $a=&$o; " + splice + " }", 0},
		{"element write", "function f(){ $a=[]; $a[\"name\"]=1; " + splice + " }", 0},
		{"element increment", "function f(){ $a=[]; @$a[\"name\"]++; " + splice + " }", 0},
		{"compound", "function f(){ $a=[]; $a+=[\"name\"=>1]; " + splice + " }", 0},
		{"destructured", "function f($o){ $a=[]; [$a]=$o; " + splice + " }", 0},
		{"passed to a call", "function f(){ $a=[]; fill($a); " + splice + " }", 0},
		{"foreach target", "function f($o){ $a=[]; foreach($o as $a[\"name\"]){} " + splice + " }", 0},
		{"extract", "function f($o){ $a=[]; extract($o); " + splice + " }", 0},
		{"include", "function f(){ $a=[]; include 'fill.php'; " + splice + " }", 0},
		{"eval", "function f($code){ $a=[]; eval($code); " + splice + " }", 0},
		{"later read", "function f(){ $a=[]; g(); " + splice + " return $a; }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArraySpliceDiscardsReplacementKeys"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
