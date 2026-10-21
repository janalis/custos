// Package iteratorcountchangesrequiredposition implements IteratorCountChangesRequiredPosition.
package iteratorcountchangesrequiredposition

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Restore the iterator position after counting."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IteratorCountChangesRequiredPosition" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ArrayIterator", "current") && !semanticquery.NativeMethod(ctx, c, "ArrayIterator", "key") {
		return
	}
	receiver, ok := syntax.UnwrapParens(c.Var).(*syntax.Variable)
	if !ok || receiver.NameExpr != nil {
		return
	}
	construct := semanticquery.NativeConstruction(ctx, c.Var, "ArrayIterator")
	if construct == nil {
		return
	}
	// Counting an empty iterator leaves the same invalid position it started at.
	if items := semanticquery.NativeArray(ctx, semanticquery.CallArgument(construct.Args, 0, "array")); items == nil || len(items.Items) == 0 {
		return
	}
	var counted *syntax.FuncCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if ok && semanticquery.NativeDominates(p, c) && semanticquery.NativeBuiltin(ctx, p, "iterator_count") && semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "iterator"), c.Var) {
			counted = p
		}
	}
	if counted != nil && !iteratorCountPositionDisturbed(ctx, receiver.Name, construct, counted, c) {
		ctx.ReportNode(c, message)
	}
}

// iteratorCountKeepsPosition recognizes ArrayIterator calls that cannot make an
// exhausted cursor valid again. Repositioning calls qualify only when a later
// count consumes the iterator anyway.
func iteratorCountKeepsPosition(ctx *analysis.Context, c *syntax.MethodCall, beforeCount bool) bool {
	for _, method := range []string{"current", "key", "valid", "next", "count"} {
		if semanticquery.NativeMethod(ctx, c, "ArrayIterator", method) {
			return true
		}
	}
	return beforeCount && (semanticquery.NativeMethod(ctx, c, "ArrayIterator", "rewind") || semanticquery.NativeMethod(ctx, c, "ArrayIterator", "seek"))
}

// iteratorCountPositionDisturbed reports whether the iterator variable is used
// between its construction and the read in a way that may move the cursor
// after the last count or hand the object to other code. Only iterator_count
// arguments and ArrayIterator calls that keep an exhausted cursor exhausted
// are transparent; rewind() and seek() are transparent before the last count
// only. A foreach subject, a loop header, a wrapper constructor, an alias, a
// closure capture or any other mention ends the proof.
func iteratorCountPositionDisturbed(ctx *analysis.Context, name string, construct *syntax.New, counted *syntax.FuncCall, read *syntax.MethodCall) bool {
	scope := syntax.EnclosingVariableScope(read)
	transparent := map[syntax.Node]bool{}
	disturbed := false
	visit := func(n syntax.Node) bool {
		if disturbed {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			// Captured objects stay shared with the enclosing scope.
		case *syntax.FuncCall:
			if semanticquery.NativeBuiltin(ctx, x, "iterator_count") {
				if arg := semanticquery.CallArgument(x.Args, 0, "iterator"); arg != nil {
					transparent[syntax.UnwrapParens(arg)] = true
				}
			}
		case *syntax.MethodCall:
			if iteratorCountKeepsPosition(ctx, x, x.Span().Start < counted.Span().Start) {
				transparent[syntax.UnwrapParens(x.Var)] = true
			}
		case *syntax.Variable:
			if x.NameExpr == nil && x.Name == name && x.Span().Start >= construct.Span().End && x.Span().Start < read.Span().Start && !transparent[n] {
				disturbed = true
			}
		default:
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return disturbed
}
