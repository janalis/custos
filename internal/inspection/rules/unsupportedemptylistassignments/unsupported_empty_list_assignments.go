package unsupportedemptylistassignments

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// unsupportedEmptyListAssignments reports foreach loops destructuring into a
// pattern without any target, a fatal error since PHP 7.0.
type unsupportedEmptyListAssignments struct{}

func (unsupportedEmptyListAssignments) ID() string { return "UnsupportedEmptyListAssignments" }
func (unsupportedEmptyListAssignments) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach}
}

func (unsupportedEmptyListAssignments) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 { // E1
		return
	}
	fe := n.(*syntax.Foreach)
	if fe.Key != nil { // E3
		return
	}
	var span syntax.Span
	switch v := fe.Value.(type) { // D2
	case *syntax.List:
		if astquery.PatternHasTarget(v.Items) { // E2
			return
		}
		span = syntax.Span{Start: v.Span().Start, End: v.Span().Start + 4}
	case *syntax.Array:
		if !v.Short || astquery.PatternHasTarget(v.Items) { // E2
			return
		}
		span = syntax.Span{Start: v.Span().Start, End: v.Span().Start + 1}
	default:
		return
	}
	ctx.Report(span, "Empty destructuring pattern: PHP 7+ rejects this with a fatal error.") // D3
}
