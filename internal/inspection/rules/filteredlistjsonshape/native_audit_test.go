package filteredlistjsonshape

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditExplicitIntent(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FilteredListJsonShape"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{"json_encode(array_filter([0,1]),JSON_FORCE_OBJECT);", "json_encode(array_filter([0,1]),$unknownFlags);"} {
		got := e.Analyze(syntax.Parse("audit.php", []byte("<?php "+src), syntax.Options{}))
		if len(got) != 0 {
			t.Fatalf("explicit intent produced %+v", got)
		}
	}
}
