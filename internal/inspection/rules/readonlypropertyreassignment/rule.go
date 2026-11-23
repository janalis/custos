// Package readonlypropertyreassignment implements the native
// ReadonlyPropertyReassignment inspection.
package readonlypropertyreassignment

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
)

type rule struct{}

// New returns the ReadonlyPropertyReassignment inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadonlyPropertyReassignment" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const message = "Initialize this readonly property only once."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.Below(phpversion.PHP81) {
		return
	}
	assign := n.(*syntax.Assign)
	// `??=` writes only when the stored value is null, which is never proven here.
	if assign.Op.Kind == syntax.TCoalesceEqual {
		return
	}
	fetch, ok := assign.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	name, ok := fetch.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	classes := ctx.TypeOf(fetch.Var).Classes()
	if len(classes) != 1 {
		return
	}
	property := ctx.Index().FindProperty(strings.TrimPrefix(classes[0], `\`), name.Value, ctx.PHP)
	if property == nil || !property.Readonly {
		return
	}
	method, ok := syntax.EnclosingFuncLike(assign).(*syntax.Method)
	if ok && strings.EqualFold(method.Name.Value, "__clone") && ctx.PHP.AtLeast(phpversion.PHP83) {
		return
	}
	if initializedByDefault(ctx, assign, strings.TrimPrefix(classes[0], `\`), property) {
		ctx.ReportNode(fetch, message)
		return
	}
	statement, ok := assign.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	_, ok = statement.Parent().(*syntax.Block)
	if !ok {
		return
	}
	priorStatements := flowquery.NativePriorStatements(ctx.File, assign)
	if len(priorStatements) == 0 {
		return
	}
	expression, ok := priorStatements[len(priorStatements)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	prior, ok := expression.Expr.(*syntax.Assign)
	if !ok || prior.ByRef {
		return
	}
	if _, ok := syntax.UnwrapParens(fetch.Var).(*syntax.Variable); !ok {
		return
	}
	// Unknown evaluation can replace the receiver or unset its property.
	switch syntax.UnwrapParens(prior.Value).(type) {
	case *syntax.Literal, *syntax.Variable:
	default:
		return
	}
	if astquery.Equivalent(ctx.File, prior.Var, assign.Var) {
		ctx.ReportNode(fetch, message)
	}
}

// initializedByDefault reports whether the property is initialized by its
// declared default value (PHP 8.6), so that no write may follow, wherever it
// is. A subclass may redeclare a non-private property without the default:
// only a private property written in its own class, or one of a final
// class, is settled. Elsewhere a private name is another property.
func initializedByDefault(ctx *analysis.Context, at syntax.Node, class string, property *index.Property) bool {
	if ctx.PHP.Below(phpversion.PHP86) || !property.HasDefault || property.Promoted {
		return false
	}
	if property.Visibility == index.Private {
		scope := syntax.EnclosingClass(at)
		return scope != nil && strings.EqualFold(strings.TrimPrefix(ctx.Types().ClassFQN(scope), `\`), strings.TrimPrefix(property.Class, `\`))
	}
	c := ctx.Index().Class(class, ctx.PHP)
	return c != nil && c.Final
}

func (rule) Semantic() {}
