package suspicioussemicolon

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// suspiciousSemicolon reports a lone `;` used as the whole body of an
// if/elseif/else clause or a loop.
type suspiciousSemicolon struct{}

const suspiciousSemicolonMsg = "This ';' is the entire body of the statement; probably unintended."

func (suspiciousSemicolon) ID() string { return "SuspiciousSemicolon" }
func (suspiciousSemicolon) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KNop}
}

func (suspiciousSemicolon) Check(ctx *analysis.Context, n syntax.Node) {
	s := n.Span()
	if s.Len() != 1 || ctx.Src[s.Start] != ';' { // D1 (skip recovery nodes)
		return
	}
	var body syntax.Stmt
	switch p := n.Parent().(type) { // D2
	case *syntax.If:
		body = p.Body
	case *syntax.ElseIf:
		body = p.Body
	case *syntax.Else:
		body = p.Body
	case *syntax.While:
		body = p.Body
	case *syntax.DoWhile:
		body = p.Body
	case *syntax.For:
		body = p.Body
	case *syntax.Foreach:
		body = p.Body
	default:
		return
	}
	if body == n.(syntax.Stmt) {
		ctx.Report(s, suspiciousSemicolonMsg)
	}
}
