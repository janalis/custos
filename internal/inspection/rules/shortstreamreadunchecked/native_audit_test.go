package shortstreamreadunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditGuardInvalidation(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ShortStreamReadUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("audit.php", []byte("<?php $v=fread($h,4);if(strlen($v)===4){$v=fread($h,4);unpack('Nsize',$v);}"), syntax.Options{}))
	if len(got) != 1 {
		t.Fatalf("new result wrongly inherits guard: %+v", got)
	}
}
