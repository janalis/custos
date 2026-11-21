package unavailablefunction

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/stubs"
)

// TestAssumedTarget: without a configured target version, the fallback
// says nothing about where the code runs, so nothing is reported.
func TestAssumedTarget(t *testing.T) {
	ix := index.New(stubs.Index())
	ix.SetWorld(index.World{Complete: true})
	f := syntax.Parse("a.php", []byte("<?php array_first([1]); each($a);"), syntax.Options{})
	// The same version configured reports both calls.
	for cfg, want := range map[phpversion.Version]int{0: 0, phpversion.PHP84: 2} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: cfg})
		if err != nil {
			t.Fatal(err)
		}
		if findings := e.WithIndex(ix).Analyze(f); len(findings) != want {
			t.Errorf("PHP %v: %d findings, want %d: %+v", cfg, len(findings), want, findings)
		}
	}
}
