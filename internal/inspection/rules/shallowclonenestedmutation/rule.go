// Package shallowclonenestedmutation implements the ShallowCloneNestedMutation inspection.
package shallowclonenestedmutation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ShallowCloneNestedMutation" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const message = "Clone the nested object before mutating independent state."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	assign := n.(*syntax.Assign)
	outer, ok := assign.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	child, ok := outer.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	name, fixedName := child.Name.(*syntax.Identifier)
	if !fixedName {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, n)
	if len(prior) < 2 {
		return
	}
	cloneStatement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	cloneAssign, ok := cloneStatement.Expr.(*syntax.Assign)
	if !ok || !astquery.Equivalent(ctx.File, cloneAssign.Var, child.Var) {
		return
	}
	value, ok := syntax.UnwrapParens(cloneAssign.Value).(*syntax.Clone)
	if !ok || value.With != nil {
		return
	}
	classes := ctx.TypeOf(value.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	final := false
	for _, cls := range classes {
		cls = strings.TrimPrefix(cls, `\`)
		decl := ctx.Index().Class(cls, ctx.PHP)
		if decl == nil {
			return
		}
		final = decl.Final
		if !ctx.Index().AncestorsComplete(cls, ctx.PHP) || !semanticquery.HierarchyResolved(ctx.Index(), cls, ctx.PHP) || ctx.Index().FindMethod(cls, "__clone", ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__get", ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__set", ctx.PHP) != nil {
			return
		}
		if property := ctx.Index().FindProperty(cls, name.Value, ctx.PHP); property != nil && (property.ReadsRunCode || property.WritesRunCode || property.Magic) {
			return
		}
	}

	statement, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	previous, ok := statement.Expr.(*syntax.Assign)
	if !ok {
		return
	}
	property, ok := previous.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	if !astquery.Equivalent(ctx.File, property.Var, value.Expr) || !astquery.Equivalent(ctx.File, property.Name, child.Name) {
		return
	}
	if _, ok := semanticquery.NativeValue(ctx, previous.Value).(*syntax.New); !ok {
		return
	}
	// A declared type admits subclasses whose __clone may deep-copy; only a
	// final class or a locally constructed instance fixes the runtime class.
	if !final && !exactInstance(ctx.File, prior[:len(prior)-2], value.Expr) {
		return
	}
	ctx.ReportNode(outer, message)
}

// exactInstance reports whether target is a variable last assigned, in the
// straight-line statements before the nested store, a `new` naming its class
// statically, so no subclass can stand in for the inferred type. Only literal
// property stores on the same variable may sit in between.
func exactInstance(f *syntax.File, prior []syntax.Stmt, target syntax.Expr) bool {
	for i := len(prior) - 1; i >= 0; i-- {
		statement, ok := prior[i].(*syntax.ExprStmt)
		if !ok {
			return false
		}
		assign, ok := statement.Expr.(*syntax.Assign)
		if !ok || assign.ByRef || assign.Op.Kind != syntax.TEqual {
			return false
		}
		if _, variable := assign.Var.(*syntax.Variable); variable && astquery.Equivalent(f, assign.Var, target) {
			created, ok := syntax.UnwrapParens(assign.Value).(*syntax.New)
			if !ok {
				return false
			}
			name, ok := created.Class.(*syntax.Name)
			return ok && !strings.EqualFold(name.Value, "static")
		}
		store, ok := assign.Var.(*syntax.PropertyFetch)
		if !ok || !astquery.Equivalent(f, store.Var, target) {
			return false
		}
		if _, literal := assign.Value.(*syntax.Literal); !literal {
			return false
		}
	}
	return false
}

func (rule) Semantic() {}
