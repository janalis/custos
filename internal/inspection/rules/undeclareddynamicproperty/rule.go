// Package undeclareddynamicproperty implements the UndeclaredDynamicProperty inspection.
package undeclareddynamicproperty

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UndeclaredDynamicProperty" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const message = "Declare this property before assigning it."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP82 {
		return
	}
	assign := n.(*syntax.Assign)
	property, ok := assign.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	name, ok := property.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	classes := ctx.TypeOf(property.Var).Classes()
	if len(classes) != 1 {
		return
	}
	cls := strings.TrimPrefix(classes[0], `\`)
	if ctx.Index().Class(cls, ctx.PHP) == nil || !ctx.Index().AncestorsComplete(cls, ctx.PHP) || ctx.Index().FindProperty(cls, name.Value, ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__set", ctx.PHP) != nil || ctx.Index().IsSubtype(cls, "stdClass", ctx.PHP) {
		return
	}
	for _, parent := range ctx.Index().Ancestors(cls, ctx.PHP) {
		for _, name := range append(append([]string{}, parent.Traits...), parent.Parent) {
			if name != "" && ctx.Index().Class(name, ctx.PHP) == nil {
				return
			}
		}
		if parent.HasAttr("AllowDynamicProperties") {
			return
		}
	}
	if class := ctx.Index().Class(cls, ctx.PHP); !class.Final {
		fresh := syntax.UnwrapParens(property.Var)
		if _, exact := fresh.(*syntax.New); !exact {
			prior := flowquery.NativePriorStatements(ctx.File, assign)
			if len(prior) == 0 {
				return
			}
			statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
			if !ok {
				return
			}
			allocation, ok := statement.Expr.(*syntax.Assign)
			if !ok || allocation.ByRef || !astquery.Equivalent(ctx.File, allocation.Var, property.Var) {
				return
			}
			fresh = syntax.UnwrapParens(allocation.Value)
		}
		allocation, exact := fresh.(*syntax.New)
		if !exact {
			return
		}
		if name, ok := allocation.Class.(*syntax.Name); !ok || strings.EqualFold(name.Value, "static") {
			return
		}
	}
	if guardedPropertyPresent(ctx, assign, property, name.Value) {
		return
	}
	ctx.ReportNode(property, message)
}

func guardedPropertyPresent(ctx *analysis.Context, at *syntax.Assign, property *syntax.PropertyFetch, name string) bool {
	// Limit the lexical guard proof to an immediately following simple write.
	switch syntax.UnwrapParens(at.Value).(type) {
	case *syntax.Literal, *syntax.Variable:
	default:
		return false
	}
	prior := flowquery.NativePriorStatements(ctx.File, at)
	if len(prior) == 0 {
		return false
	}
	guard, ok := prior[len(prior)-1].(*syntax.If)
	if !ok || guard.Else != nil || len(guard.ElseIfs) != 0 {
		return false
	}
	negative, ok := syntax.UnwrapParens(guard.Cond).(*syntax.Unary)
	if !ok || negative.Op.Kind != syntax.TExclaim {
		return false
	}
	call, _ := semanticquery.GlobalCall(ctx, syntax.UnwrapParens(negative.Expr), "property_exists")
	if call == nil || !astquery.Equivalent(ctx.File, semanticquery.CallArgument(call.Args, 0, "object_or_class"), property.Var) {
		return false
	}
	field, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "property"))
	if !known || field != name {
		return false
	}
	stmts, listed := syntax.StmtListOf(guard.Body)
	if !listed || len(stmts) != 1 {
		return false
	}
	switch stmt := stmts[0].(type) {
	case *syntax.Return:
		return true
	case *syntax.ExprStmt:
		_, terminal := stmt.Expr.(*syntax.Throw)
		return terminal
	}
	return false
}

func (rule) Semantic() {}
