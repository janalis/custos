package arrowfunctioncapturedcounter

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	v "custos/internal/php/version"
)

func TestBeforeLanguageFeature(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrowFunctionCapturedCounter"}, PHP: v.PHP73})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("old.php", []byte("<?php $n=0; $f=fn()=>++$n; echo $f(),$f();"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("old version produced findings: %+v", got)
	}
}
