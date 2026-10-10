package validatedbooleanfalserejectedasinvalid

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
		{"<?php\nif (filter_var('false', FILTER_VALIDATE_BOOLEAN) === false) { throw new InvalidArgumentException(); }\n", 1},
		{"<?php\nif (filter_var('false', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === null) { throw new InvalidArgumentException(); }\n", 0},
		{"<?php other();", 0},
		{"<?php if(filter_var('true',FILTER_VALIDATE_BOOLEAN)===false){return false;}", 0},
		{"<?php if(filter_var($s,FILTER_VALIDATE_BOOLEAN)===false){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN,$options)===false){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN,FILTER_NULL_ON_FAILURE)===false){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_INT)===false){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN)===true){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN)===false){echo 'false';}", 0},
		{"<?php $v=filter_var('false',FILTER_VALIDATE_BOOLEAN)===false;", 0},
		{"<?php if(filter_var('off',FILTER_VALIDATE_BOOLEAN)===false){return false;}", 1},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN)===0){return false;}", 0},
		{"<?php if(filter_var('false',FILTER_VALIDATE_BOOLEAN)==false){return false;}", 0},
		{"<?php if($v===false){return false;}", 0},
		{"<?php function unreachable(){return;\nif (filter_var('false', FILTER_VALIDATE_BOOLEAN) === false) { throw new InvalidArgumentException(); }\n}", 0},
		{"<?php \nif (filter_var('false', FILTER_VALIDATE_BOOLEAN) === false) { throw new InvalidArgumentException(); }\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ValidatedBooleanFalseRejectedAsInvalid"}})
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
	for _, name := range []string{"basic", "negative"} {
		marked, err := os.ReadFile("../../../../testdata/rules/ValidatedBooleanFalseRejectedAsInvalid/" + name + ".php")
		if err != nil {
			t.Fatal(err)
		}
		src, want, err := conformance.ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ValidatedBooleanFalseRejectedAsInvalid"}})
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
