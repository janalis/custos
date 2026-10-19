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

type queryGapRule struct{ typeGapRule }

func (queryGapRule) Check(ctx *Context, n syntax.Node) {
	for i := range MaxGaps + 2 {
		ctx.Gap("query-budget", syntax.Span{Start: uint32(i), End: uint32(i + 1)})
		ctx.Gap("query-budget", syntax.Span{Start: uint32(i), End: uint32(i + 1)})
	}
}

// Query gaps are recorded once per reason and range, and at most MaxGaps.
func TestQueryGapsAreBounded(t *testing.T) {
	engine, err := NewEngine([]Rule{queryGapRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	result := engine.AnalyzeResult(syntax.Parse("query.php", []byte(`<?php f();`), syntax.Options{}))
	if len(result.Incomplete) != MaxGaps || result.Incomplete[0].Reason != "query-budget" || result.Incomplete[1].Span.Start != 1 {
		t.Fatalf("%d gaps: %+v", len(result.Incomplete), result.Incomplete[:2])
	}
}
