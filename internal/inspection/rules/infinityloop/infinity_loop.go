package infinityloop

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// infinityLoop reports a method whose only statement calls itself on
// $this/self/static.
type infinityLoop struct{}

const infinityLoopMsg = "Method calls itself unconditionally; this recursion never ends."

func (infinityLoop) ID() string               { return "InfinityLoop" }
func (infinityLoop) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }
func (infinityLoop) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	if m.Body == nil || m.Name == nil || len(m.Body.Stmts) != 1 { // D1, D2
		return
	}
	var cand syntax.Expr // D3
	switch s := m.Body.Stmts[0].(type) {
	case *syntax.Return:
		cand = s.Expr
	case *syntax.ExprStmt:
		cand = s.Expr
	}
	var name, recv syntax.Expr
	switch c := cand.(type) { // D4
	case *syntax.MethodCall:
		name, recv = c.Name, c.Var
	case *syntax.StaticCall:
		name, recv = c.Name, c.Class
	default:
		return
	}
	id, ok := name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, m.Name.Value) {
		return
	}
	switch t := ctx.Text(recv); { // D5
	case t == "$this", strings.EqualFold(t, "self"), strings.EqualFold(t, "static"):
		ctx.ReportNode(cand, infinityLoopMsg)
	}
}
