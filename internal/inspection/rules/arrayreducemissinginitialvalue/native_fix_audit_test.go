package arrayreducemissinginitialvalue

import (
	"bytes"
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditFixBytes(t *testing.T) {
	src := []byte("<?php array_reduce($items,fn($carry,$v)=>array_merge($carry,[$v]));")
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayReduceMissingInitialValue"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("before.php", src, syntax.Options{}))
	if len(got) != 1 || len(got[0].Fixes) != 1 {
		t.Fatalf("findings %+v", got)
	}
	fixed, _ := fixing.Apply(src, got[0].Fixes[0].Edits())
	want := []byte("<?php array_reduce($items,fn($carry,$v)=>array_merge($carry,[$v]), []);")
	if !bytes.Equal(fixed, want) {
		t.Fatalf("fixed %s, want %s", fixed, want)
	}
	if again := e.Analyze(syntax.Parse("after.php", fixed, syntax.Options{})); len(again) != 0 {
		t.Fatalf("fix did not resolve findings %+v", again)
	}
}
