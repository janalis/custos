package filereadfailureunchecked

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/flow"
)

func TestAssertionAssumption(t *testing.T) {
	src := []byte(`<?php function f(string $p){$v=file_get_contents($p);assert($v!==false);trim($v);}`)
	for _, enabled := range []bool{false, true} {
		engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileReadFailureUnchecked"}, Flow: flow.Options{Assertions: enabled}})
		if err != nil {
			t.Fatal(err)
		}
		got := engine.Analyze(syntax.Parse("test.php", src, syntax.Options{}))
		want := 1
		if enabled {
			want = 0
		}
		if len(got) != want {
			t.Fatalf("enabled=%v findings=%v", enabled, got)
		}
	}
}
