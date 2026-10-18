package analysis

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

type typeGapRule struct{}

func (typeGapRule) ID() string                        { return "NestedNotOperators" }
func (typeGapRule) Semantic()                         {}
func (typeGapRule) Kinds() []syntax.NodeKind          { return []syntax.NodeKind{syntax.KFuncCall} }
func (typeGapRule) Check(ctx *Context, n syntax.Node) { ctx.TypeOf(n.(syntax.Expr)) }

func TestDetailedAnalysisIncludesTypeBudgetsWithoutFlow(t *testing.T) {
	engine, err := NewEngine([]Rule{typeGapRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	file := syntax.Parse("query.php", []byte(`<?php max(`+strings.Repeat("1,", 256)+`1);`), syntax.Options{})
	result := engine.AnalyzeResult(file)
	if len(result.Findings) != 0 || len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "argument-budget" {
		t.Fatal(result)
	}
}

type flowGapRule struct{ typeGapRule }

func (flowGapRule) Flow() {}
func (flowGapRule) Check(ctx *Context, n syntax.Node) {
	ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
}

func TestDetailedAnalysisIncludesFlowAndParserGaps(t *testing.T) {
	for _, tc := range []struct{ source, reason string }{
		{"<?php missingCall();", "unresolved-dispatch"},
		{"<?php function {", "syntax-recovery"},
	} {
		engine, err := NewEngine([]Rule{flowGapRule{}}, Config{EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		file := syntax.Parse("gaps.php", []byte(tc.source), syntax.Options{})
		result := engine.AnalyzeResult(file)
		if len(result.Incomplete) != 1 || result.Incomplete[0].Reason != tc.reason {
			t.Fatal(result)
		}
	}
}
