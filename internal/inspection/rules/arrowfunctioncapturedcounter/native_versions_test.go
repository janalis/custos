package arrowfunctioncapturedcounter

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

func TestBeforeLanguageFeature(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrowFunctionCapturedCounter"}, PHP: phpversion.PHP73})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("old.php", []byte("<?php $n=0; $f=fn()=>++$n; echo $f(),$f();"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("old version produced findings: %+v", got)
	}
}
