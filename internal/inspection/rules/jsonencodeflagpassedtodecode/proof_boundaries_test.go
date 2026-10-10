package jsonencodeflagpassedtodecode

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "JsonEncodeFlagPassedToDecode" || len(r.Kinds()) == 0 {
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"JsonEncodeFlagPassedToDecode"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $a=json_decode($json,true,512,JSON_PRETTY_PRINT);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"json_encode([]);", 0}, {"json_decode($x,true,512,JSON_BIGINT_AS_STRING|0);", 0}, {"json_decode($x,true,512,JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING|JSON_BIGINT_AS_STRING);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonEncodeFlagPassedToDecode"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
