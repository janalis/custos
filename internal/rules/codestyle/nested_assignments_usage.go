package codestyle

import (
	"custos/internal/analysis/util"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// nestedAssignmentsUsage reports chained assignments `$a = $b = value`.
type nestedAssignmentsUsage struct{}

func init() { register(nestedAssignmentsUsage{}) }

func (nestedAssignmentsUsage) ID() string { return "NestedAssignmentsUsage" }

func (nestedAssignmentsUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const nestedAssignmentsMsg = "Split this chained assignment into separate assignments."

func (nestedAssignmentsUsage) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Assign)
	if a.Op.Kind != syntax.TEqual || naDestructuring(a.Var) { // E2, E3
		return
	}
	if _, ok := a.Value.(*syntax.Assign); !ok { // D1, E1
		return
	}
	if _, ok := a.Parent().(*syntax.Assign); ok { // D2
		return
	}
	stmt, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		ctx.ReportNode(a, nestedAssignmentsMsg)
		return
	}
	// Collect the chain V1 = ... = Vn = E (F4: plain links only).
	var targets []syntax.Expr
	var e syntax.Expr = a
	for {
		link, ok := e.(*syntax.Assign)
		if !ok {
			break
		}
		if link.Op.Kind != syntax.TEqual || link.ByRef || naDestructuring(link.Var) {
			ctx.ReportNode(a, nestedAssignmentsMsg)
			return
		}
		targets = append(targets, link.Var)
		e = link.Value
	}
	src := ctx.Src
	span := stmt.Span()
	simple := naSimpleValue(e)
	ctx.ReportNode(a, nestedAssignmentsMsg, analysis.Fix{
		Title: "Split into separate assignments",
		Edits: func() []analysis.TextEdit {
			text := func(n syntax.Node) string { s := n.Span(); return string(src[s.Start:s.End]) }
			indent := util.IndentBefore(src, span.Start)
			last := text(targets[len(targets)-1])
			value := text(e)
			var b strings.Builder
			for i := len(targets) - 1; i >= 0; i-- {
				if i != len(targets)-1 {
					b.WriteString("\n" + indent)
				}
				v := value
				if !simple && i != len(targets)-1 {
					v = last
				}
				b.WriteString(text(targets[i]) + " = " + v + ";")
			}
			return []analysis.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
}

func naDestructuring(e syntax.Expr) bool {
	switch e.(type) {
	case *syntax.Array, *syntax.List:
		return true
	}
	return false
}

// naSimpleValue implements F1.
func naSimpleValue(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Variable, *syntax.ConstFetch, *syntax.ClassConstFetch, *syntax.Literal:
		return true
	case *syntax.Unary:
		l, ok := e.Expr.(*syntax.Literal)
		return e.Op.Kind == syntax.TMinus && ok && l.LitKind != syntax.LitString
	}
	return false
}
