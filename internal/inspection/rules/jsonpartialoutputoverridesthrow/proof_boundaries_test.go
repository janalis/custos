package jsonpartialoutputoverridesthrow

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "JsonPartialOutputOverridesThrow" || len(r.Kinds()) == 0 {
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"JsonPartialOutputOverridesThrow"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $s=json_encode($data,JSON_PARTIAL_OUTPUT_ON_ERROR|JSON_THROW_ON_ERROR);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"json_decode(\"{}\");", 0}, {"json_encode($x,JSON_THROW_ON_ERROR|$unknown);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonPartialOutputOverridesThrow"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
