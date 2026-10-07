package confusing

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// referencingObjects reports objects taken by reference: typed by-ref
// parameters that are never re-assigned, and `= &new`.
type referencingObjects struct{}

func init() { register(referencingObjects{}) }

func (referencingObjects) ID() string { return "ReferencingObjects" }

func (referencingObjects) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KParam, syntax.KNew}
}

func (r referencingObjects) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Param:
		r.checkParam(ctx, x)
	case *syntax.New:
		r.checkNew(ctx, x)
	}
}

func (referencingObjects) checkParam(ctx *analysis.Context, p *syntax.Param) {
	var body *syntax.Block
	switch fn := p.Parent().(type) { // D1
	case *syntax.Function:
		body = fn.Body
	case *syntax.Method:
		body = fn.Body
	default:
		return
	}
	if !p.ByRef || p.Default != nil || p.Var.NameExpr != nil || p.Type == nil { // D2/D3
		return
	}
	if scalarOnlyType(p.Type) {
		return
	}
	if body != nil && usedAsReference(body, p.Var.Name) { // D4
		return
	}
	f := ctx.File
	span := syntax.Span{Start: p.Span().Start, End: p.Var.Span().End}
	gap := syntax.Span{Start: p.Type.Span().End, End: p.Var.Span().Start}
	if p.Variadic {
		if t, ok := util.FindToken(f, gap, syntax.TEllipsis); ok {
			gap.End = t.Start
		}
	}
	ctx.Report(span, "Objects are handed over by handle already; drop the '&' before '$"+p.Var.Name+"'.", analysis.Fix{
		Title: "Remove the '&'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: gap, NewText: " "}} },
	})
}

var scalarTypeNames = map[string]bool{
	"string": true, "int": true, "float": true, "bool": true, "array": true,
	"mixed": true, "iterable": true, "null": true,
	"callable": true, // may be a string or an array (spec Divergences)
}

// scalarOnlyType reports whether every name of the type is a scalar/array
// pseudo-type (D3 fails).
func scalarOnlyType(t syntax.Expr) bool {
	switch x := t.(type) {
	case *syntax.Name:
		return scalarTypeNames[strings.ToLower(strings.TrimPrefix(x.Value, `\`))]
	case *syntax.NullableType:
		return scalarOnlyType(x.Type)
	case *syntax.UnionType:
		for _, m := range x.Types {
			if !scalarOnlyType(m) {
				return false
			}
		}
		return true
	}
	return false // intersection members are always classes
}

// usedAsReference reports whether the parameter is assigned to or used as a
// logical operand somewhere in body.
func usedAsReference(body *syntax.Block, name string) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if found {
			return false
		}
		v, ok := n.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name != name {
			return true
		}
		if a, ok := v.Parent().(*syntax.Assign); ok && a.Var == syntax.Expr(v) {
			found = true
			return false
		}
		found = util.IsLogicalOperand(v)
		return !found
	})
	return found
}

func (referencingObjects) checkNew(ctx *analysis.Context, nw *syntax.New) {
	a, ok := nw.Parent().(*syntax.Assign) // D5
	if !ok || !a.ByRef || a.Value != syntax.Expr(nw) {
		return
	}
	gap := syntax.Span{Start: a.Op.Span.Start, End: nw.Span().Start}
	ctx.ReportNode(nw, "Objects are handed over by handle already; assign the new instance without '&'.", analysis.Fix{
		Title: "Remove the '&'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: gap, NewText: "= "}} },
	})
}
