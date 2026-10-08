package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
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
	// custos: the outer targets read the innermost one back; that is only
	// equivalent when reading it has no side effect and yields the stored
	// value (`$t = $list[] = f()` cannot read `$list[]`).
	if !simple && !naRereadable(targets[len(targets)-1]) {
		ctx.ReportNode(a, nestedAssignmentsMsg)
		return
	}
	ctx.ReportNode(a, nestedAssignmentsMsg, analysis.Fix{
		Title: "Split into separate assignments",
		Edits: func() []analysis.TextEdit {
			text := func(n syntax.Node) string { s := n.Span(); return string(src[s.Start:s.End]) }
			indent := util.IndentBefore(src, span.Start)
			last := text(targets[len(targets)-1])
			value := text(e)
			// custos: a brace-less control body (`if ($c) $a = $b = 1;
			// else …`) needs braces around the split statements.
			braces := false
			switch stmt.Parent().(type) {
			case *syntax.If, *syntax.ElseIf, *syntax.Else, *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Declare:
				braces = true
			}
			sep := indent
			var b strings.Builder
			if braces {
				sep = indent + "    "
				b.WriteString("{\n" + sep)
			}
			for i := len(targets) - 1; i >= 0; i-- {
				if i != len(targets)-1 {
					b.WriteString("\n" + sep)
				}
				v := value
				if !simple && i != len(targets)-1 {
					v = last
				}
				b.WriteString(text(targets[i]) + " = " + v + ";")
			}
			if braces {
				b.WriteString("\n" + indent + "}")
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

// naRereadable reports whether a target can be read back after the write:
// a plain variable, or property/element accesses on one with identifier
// names and literal, constant or plain-variable keys.
func naRereadable(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.Variable:
		return e.NameExpr == nil
	case *syntax.ArrayDimFetch:
		if e.Dim == nil {
			return false
		}
		switch d := e.Dim.(type) {
		case *syntax.Literal, *syntax.ConstFetch, *syntax.ClassConstFetch:
		case *syntax.Variable:
			if d.NameExpr != nil {
				return false
			}
		default:
			return false
		}
		return naRereadable(e.Var)
	case *syntax.PropertyFetch:
		if _, ok := e.Name.(*syntax.Identifier); !ok || e.NullSafe {
			return false
		}
		return naRereadable(e.Var)
	case *syntax.StaticPropertyFetch:
		_, cls := e.Class.(*syntax.Name)
		v, ok := e.Name.(*syntax.Variable)
		return cls && ok && v.NameExpr == nil
	}
	return false
}
