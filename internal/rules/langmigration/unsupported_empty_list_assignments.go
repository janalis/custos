package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// unsupportedEmptyListAssignments reports foreach loops destructuring into a
// pattern without any target, a fatal error since PHP 7.0.
type unsupportedEmptyListAssignments struct{}

func init() { register(unsupportedEmptyListAssignments{}) }

func (unsupportedEmptyListAssignments) ID() string { return "UnsupportedEmptyListAssignments" }

func (unsupportedEmptyListAssignments) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach}
}

func (unsupportedEmptyListAssignments) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP70 { // E1
		return
	}
	fe := n.(*syntax.Foreach)
	if fe.Key != nil { // E3
		return
	}
	var span syntax.Span
	switch v := fe.Value.(type) { // D2
	case *syntax.List:
		if patternHasTarget(v.Items) { // E2
			return
		}
		span = syntax.Span{Start: v.Span().Start, End: v.Span().Start + 4}
	case *syntax.Array:
		if !v.Short || patternHasTarget(v.Items) { // E2
			return
		}
		span = syntax.Span{Start: v.Span().Start, End: v.Span().Start + 1}
	default:
		return
	}
	if span.End > n.Span().End {
		return
	}
	ctx.Report(span, "Empty destructuring pattern: PHP 7+ rejects this with a fatal error.") // D3
}
