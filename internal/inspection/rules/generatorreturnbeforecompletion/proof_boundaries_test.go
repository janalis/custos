package generatorreturnbeforecompletion

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "GeneratorReturnBeforeCompletion" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestUnavailableVersion(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"GeneratorReturnBeforeCompletion"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $g=(function(){yield 1;return 2;})(); $v=$g->getReturn();"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"$x->start();", 0}, {"$x->$name();", 0}, {"new stdClass(); $x->getReturn();", 0}, {"$x=1; $x->getReturn();", 0}, {"$x=function(){yield 1;}; $x->getReturn();", 0}, {"$x=(function(){return 1;})(); $x->getReturn();", 0}, {"$x=(function(){yield 1;})(); echo $x->getReturn();", 0}, {"if($a) $x->getReturn();", 0}, {"$g=&$other; $g->getReturn();", 0}, {"$g=(function(){yield 1;})(); $other->getReturn();", 0}, {"$obj->p=1; $obj->getReturn();", 0}, {"(new stdClass)->getReturn();", 0}, {"if($x){} $g->getReturn();", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GeneratorReturnBeforeCompletion"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
