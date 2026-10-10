package pregreplacefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditGuardInvalidation(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PregReplaceFailureUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("audit.php", []byte("<?php $v=preg_replace('/x/','y','x');if($v!==null){$v=preg_replace('/x/','y','x');trim($v);}"), syntax.Options{}))
	if len(got) != 1 {
		t.Fatalf("new result wrongly inherits guard: %+v", got)
	}
}
