// Package arrayuniqueonnonstringableobjects implements the native ArrayUniqueOnNonStringableObjects inspection.
package arrayuniqueonnonstringableobjects

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/stubs"
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
		if nonStringable(ctx, fqn) {
			ctx.ReportNode(call, message)
			return
		}
	}
}

// throwingBuiltins lists the PHP-provided classes verified to raise an error
// on string conversion (lower-case FQN). Other internal classes may convert
// through a cast handler without declaring __toString (GMP does), so a
// missing method proves nothing for them.
var throwingBuiltins = map[string]bool{
	"stdclass": true, "arrayobject": true, "arrayiterator": true,
	"datetime": true, "datetimeimmutable": true, "datetimezone": true, "dateinterval": true,
	"splobjectstorage": true, "spldoublylinkedlist": true, "splstack": true, "splqueue": true,
	"splfixedarray": true, "splpriorityqueue": true, "splminheap": true, "splmaxheap": true,
	"weakmap": true,
}

// nonStringable reports whether objects of class fqn provably cannot be
// converted to string: the whole hierarchy is known, declares no __toString
// and contains no internal class outside throwingBuiltins.
func nonStringable(ctx *analysis.Context, fqn string) bool {
	ix := ctx.Index()
	if ix.Class(fqn, ctx.PHP) == nil || ix.FindMethod(fqn, "__toString", ctx.PHP) != nil ||
		ix.IsSubtype(fqn, "Stringable", ctx.PHP) || !ix.AncestorsComplete(fqn, ctx.PHP) {
		return false
	}
	for _, c := range ix.Ancestors(fqn, ctx.PHP) {
		if c.Kind == syntax.KindInterface {
			continue
		}
		if stubs.Index().Class(c.FQN, ctx.PHP) == c && !throwingBuiltins[strings.ToLower(c.FQN)] {
			return false
		}
		if c.Parent != "" && ix.Class(c.Parent, ctx.PHP) == nil {
			return false
		}
		for _, trait := range c.Traits {
			if ix.Class(trait, ctx.PHP) == nil {
				return false
			}
		}
	}
	return true
}
