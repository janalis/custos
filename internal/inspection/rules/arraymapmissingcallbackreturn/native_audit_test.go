package arraymapmissingcallbackreturn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditExplicitIntent(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMapMissingCallbackReturn"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("audit.php", []byte("<?php array_map(function($s):void{trim($s);},$names);"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("explicit intent produced %+v", got)
	}
}

func TestAuditDeclaredResultWithholdsFix(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMapMissingCallbackReturn"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("typed.php", []byte("<?php array_map(function($s):int{trim($s);},$names);"), syntax.Options{}))
	if len(got) != 1 || len(got[0].Fixes) != 0 {
		t.Fatalf("declared result changed by fix: %+v", got)
	}
}
