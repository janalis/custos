package analysis_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/inspection/rules/redundanttypecheck"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestSharedDataFailureWithRealRules(t *testing.T) {
	restore := analysis.FailNamesForTest()
	t.Cleanup(restore)
	engine, err := analysis.NewEngine([]analysis.Rule{redundanttypecheck.New(), nestednotoperators.New()}, analysis.Config{EnableAll: true, PHP: phpversion.Max})
	if err != nil {
		t.Fatal(err)
	}
	file := syntax.Parse("failure.php", []byte(`<?php function f(int $x) { is_string($x); return !!!$x; }`), syntax.Options{Version: phpversion.Max})
	result := engine.AnalyzeResult(file)
	found := false
	for _, finding := range result.Findings {
		if finding.Rule == "NestedNotOperators" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unrelated real rule lost: %+v", result)
	}
	if len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "internal-failure" {
		t.Fatalf("missing shared failure status: %+v", result)
	}
}
