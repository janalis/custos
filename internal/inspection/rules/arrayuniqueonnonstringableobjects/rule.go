// Package arrayuniqueonnonstringableobjects implements the native ArrayUniqueOnNonStringableObjects inspection.
package arrayuniqueonnonstringableobjects

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Choose an explicit comparison for object deduplication."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayUniqueOnNonStringableObjects" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_unique") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	if flags != nil {
		f, ok := semanticquery.NativeInt(ctx, flags)
		if !ok || f != 2 {
			return
		}
	}
	a, effective := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if !effective || len(a) < 2 {
		return
	}
	for _, item := range a {
		creation, ok := semanticquery.NativeValue(ctx, item).(*syntax.New)
		if !ok {
			continue
		}
		name, ok := creation.Class.(*syntax.Name)
		if !ok {
			continue
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		class := ctx.Index().Class(fqn, ctx.PHP)
		if class == nil {
			continue
		}
		if ctx.Index().FindMethod(fqn, "__toString", ctx.PHP) == nil && !ctx.Index().IsSubtype(fqn, "Stringable", ctx.PHP) {
			ctx.ReportNode(call, message)
			return
		}
	}
}
