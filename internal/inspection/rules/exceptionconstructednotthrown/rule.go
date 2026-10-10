// Package exceptionconstructednotthrown implements the native ExceptionConstructedNotThrown inspection.
package exceptionconstructednotthrown

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const message = "Throw the discarded exception."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExceptionConstructedNotThrown" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	creation := n.(*syntax.New)
	if _, ok := creation.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	name, ok := creation.Class.(*syntax.Name)
	if !ok {
		return
	}
	fqn := ctx.Names().Class(name.Value, name.Span().Start)
	constructor := ctx.Index().FindMethod(fqn, "__construct", ctx.PHP)
	if constructor == nil || !constructor.Builtin {
		return
	}
	if destructor := ctx.Index().FindMethod(fqn, "__destruct", ctx.PHP); destructor != nil && !destructor.Builtin {
		return
	}
	if ctx.Index().IsSubtype(fqn, "Throwable", ctx.PHP) || ctx.Index().IsSubtype(fqn, "Exception", ctx.PHP) {
		ctx.ReportNode(creation, message)
	}
}
