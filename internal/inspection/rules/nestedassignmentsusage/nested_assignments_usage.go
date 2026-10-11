package nestedassignmentsusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// nestedAssignmentsUsage reports chained assignments `$a = $b = value`.
type nestedAssignmentsUsage struct{}

func (nestedAssignmentsUsage) ID() string               { return "NestedAssignmentsUsage" }
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
	// custos: a statement closed by `?>` keeps its terminator; only the
	// expression is replaced.
	if src[span.End-1] != ';' {
		span.End = a.Span().End
	}
	simple := naSimpleValue(e)
	if !naSplittable(targets, e, simple) {
		ctx.ReportNode(a, nestedAssignmentsMsg)
		return
	}
	ctx.ReportNode(a, nestedAssignmentsMsg, diagnostic.Fix{
		Title: "Split into separate assignments",
		Edits: func() []diagnostic.TextEdit {
			text := func(n syntax.Node) string { s := n.Span(); return string(src[s.Start:s.End]) }
			indent := astquery.IndentBefore(src, span.Start)
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
			return []diagnostic.TextEdit{{Span: span, NewText: b.String()}}
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

// naSplittable reports whether the split statements are provably
// equivalent to the chain (custos, see the spec's Divergences).
func naSplittable(targets []syntax.Expr, value syntax.Expr, simple bool) bool {
	last := len(targets) - 1
	// The split runs the outermost target after the inner writes, while
	// PHP evaluates a non-trivial key or receiver of it before them.
	outer := targets[0]
	if d, ok := outer.(*syntax.ArrayDimFetch); ok && d.Dim == nil {
		outer = d.Var
	}
	if !naRereadable(outer) {
		return false
	}
	// A middle target both stores and forwards the value: only a plain
	// variable forwards it unchanged.
	for _, t := range targets[1:last] {
		if _, ok := astquery.PlainVariableName(t); !ok {
			return false
		}
	}
	inner := targets[last]
	if !simple {
		// The outer targets read the innermost one back (`$t = $list[] =
		// f()` cannot read `$list[]`).
		return naRereadable(inner)
	}
	// The value is copied to every target: a typed property or a string
	// offset would have handed a converted value to the outer targets.
	if _, ok := astquery.PlainVariableName(inner); ok || syntax.IsNullConst(value) {
		return true
	}
	return naSuperglobalElement(inner)
}

// naSuperglobalElement matches `$_SESSION['k']`: a direct element of a
// superglobal array stores the value as is.
func naSuperglobalElement(e syntax.Expr) bool {
	d, ok := e.(*syntax.ArrayDimFetch)
	if !ok {
		return false
	}
	switch name, _ := astquery.PlainVariableName(d.Var); name {
	case "_SESSION", "_GET", "_POST", "_COOKIE", "_SERVER", "_ENV", "_FILES", "_REQUEST", "GLOBALS":
		return true
	}
	return false
}

// naSimpleValue implements F1.
func naSimpleValue(e syntax.Expr) bool {
	switch e := e.(type) {
	case *syntax.ConstFetch, *syntax.Literal:
		return true
	case *syntax.Variable:
		return e.NameExpr == nil
	case *syntax.ClassConstFetch:
		_, ok := e.Class.(*syntax.Name)
		return ok
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
