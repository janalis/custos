package codestyle

import (
	"custos/internal/analysis"
	"custos/internal/syntax"
)

// usingInclusionReturnValue reports include/require expressions whose value
// is used, i.e. that are not a statement of their own.
type usingInclusionReturnValue struct{}

func init() { register(usingInclusionReturnValue{}) }

const usingInclusionReturnValueMsg = "Avoid relying on the value returned by an included file."

func (usingInclusionReturnValue) ID() string { return "UsingInclusionReturnValue" }

func (usingInclusionReturnValue) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInclude}
}

func (usingInclusionReturnValue) Check(ctx *analysis.Context, n syntax.Node) {
	if n.Span().Len() == 0 {
		return
	}
	parent := n.Parent()
	for { // D2: look through `@` and parentheses
		if u, ok := parent.(*syntax.Unary); ok && u.Op.Kind == syntax.TAt {
			parent = u.Parent()
			continue
		}
		if p, ok := parent.(*syntax.Paren); ok {
			parent = p.Parent()
			continue
		}
		break
	}
	if _, ok := parent.(*syntax.ExprStmt); ok { // E1
		return
	}
	ctx.ReportNode(n, usingInclusionReturnValueMsg)
}
