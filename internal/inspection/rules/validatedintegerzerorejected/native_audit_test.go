package validatedintegerzerorejected

import (
	"strings"
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditFixPreservesValidationComment(t *testing.T) {
	src := []byte("<?php if(! /* zero is accepted */ (filter_var($v,FILTER_VALIDATE_INT))){reject();}")
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ValidatedIntegerZeroRejected"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("comment.php", src, syntax.Options{}))
	if len(got) != 1 || len(got[0].Fixes) != 1 {
		t.Fatalf("findings %+v", got)
	}
	fixed, _ := fixing.Apply(src, got[0].Fixes[0].Edits())
	if !strings.Contains(string(fixed), "/* zero is accepted */") {
		t.Fatalf("lost comment: %s", fixed)
	}
}
