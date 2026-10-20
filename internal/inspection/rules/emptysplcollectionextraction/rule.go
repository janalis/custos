// Package emptysplcollectionextraction implements the native EmptySplCollectionExtraction inspection.
package emptysplcollectionextraction

import (
	"math"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Check that the collection contains an element."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EmptySplCollectionExtraction" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	name, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	method := strings.ToLower(name.Value)
	switch method {
	case "extract", "pop", "shift", "dequeue", "top", "bottom":
	default:
		return
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 {
		return
	}
	class := strings.TrimPrefix(classes[0], `\`)
	switch strings.ToLower(class) {
	case "splqueue", "splstack", "spldoublylinkedlist", "splminheap", "splmaxheap", "splpriorityqueue":
	default:
		return
	}
	built := semanticquery.NativeConstruction(ctx, c.Var, class)
	if !semanticquery.NativeMethod(ctx, c, class, name.Value) || built == nil {
		return
	}
	if !knownHistory(ctx, c, built) {
		return
	}
	size := 0
	historyCalls, historyKnown := semanticquery.NativeCallHistory(ctx, c, c.Var, "insert", "push", "unshift", "enqueue", "extract", "pop", "shift", "dequeue").Certain()
	if !historyKnown {
		return
	}
	for _, n := range historyCalls {
		p := n.(*syntax.MethodCall)
		switch strings.ToLower(p.Name.(*syntax.Identifier).Value) {
		case "insert", "push", "unshift", "enqueue":
			size++
		default:
			if size > 0 {
				size--
			} else {
				return
			}
		}
	}
	if size == 0 && !guardedNonempty(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

// guardedNonempty reports whether an `isEmpty()` or `count()` test of the
// receiver, earlier in the scope, controls whether the call runs: the
// result of that test, narrowed by the conditions guarding the call, says
// the collection holds an element (or the call is unreachable). The proven
// history has no nested scope before the call.
func guardedNonempty(ctx *analysis.Context, c *syntax.MethodCall) bool {
	env := ctx.Types().Native()
	found := false
	scope := syntax.EnclosingVariableScope(c)
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= c.Span().Start {
			return false
		}
		test, ok := n.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeSameValue(ctx, test.Var, c.Var) {
			return true
		}
		switch {
		case sizeMethod(ctx, test, "isEmpty"):
			t := env.CallResultAt(test, c)
			found = t.IsUnknown() || t.OnlyOf("false")
		case sizeMethod(ctx, test, "count"):
			t := env.CallResultAsAt(test, c, types.Ints(types.IntRange{Lo: 0, Hi: math.MaxInt64}))
			found = t.IsUnknown() || t.IntsOnly(1, math.MaxInt64)
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}

// sizeMethod reports whether call invokes the builtin method of that name
// shared by the collections.
func sizeMethod(ctx *analysis.Context, call *syntax.MethodCall, method string) bool {
	return semanticquery.NativeMethod(ctx, call, "SplDoublyLinkedList", method) || semanticquery.NativeMethod(ctx, call, "SplHeap", method) || semanticquery.NativeMethod(ctx, call, "SplPriorityQueue", method)
}

// Only an entirely modeled history can establish an empty collection.
func knownHistory(ctx *analysis.Context, at *syntax.MethodCall, built *syntax.New) bool {
	valid, budget := true, 256
	scope := syntax.EnclosingVariableScope(at)
	// A loop that does not rebuild the collection replays its whole body before
	// the call, so the history extends to the end of the outermost such loop.
	limit := at.Span().Start
	for p := at.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			if !p.Span().Contains(built.Span()) {
				limit = p.Span().End
			}
		}
	}
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			valid = false
			return false
		}
		if n.Span().Start >= limit {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			valid = false
			return false
		}
		switch n := n.(type) {
		case *syntax.Label:
			// A jump target can replay everything that follows it.
			limit = ^uint32(0)
		case *syntax.Variable:
			if !semanticquery.NativeSameValue(ctx, n, at.Var) {
				return true
			}
			// Any use other than a method receiver or a plain rebinding may
			// hand the collection to code that fills it.
			modeled := false
			switch parent := n.Parent().(type) {
			case *syntax.MethodCall:
				modeled = parent.Var == n
			case *syntax.Assign:
				modeled = parent.Var == n && !parent.ByRef
			}
			if !modeled {
				valid = false
			}
		case *syntax.MethodCall:
			if n == at || !semanticquery.NativeSameValue(ctx, n.Var, at.Var) {
				return true
			}
			name := ""
			if id, ok := n.Name.(*syntax.Identifier); ok {
				name = strings.ToLower(id.Value)
			}
			replayed := n.Span().Start >= at.Span().Start
			switch name {
			case "insert", "push", "unshift", "enqueue":
				if replayed || !semanticquery.NativeDominates(n, at) {
					valid = false
				}
			case "extract", "pop", "shift", "dequeue":
				// A later removal can only shrink the collection.
				if !replayed && !semanticquery.NativeDominates(n, at) {
					valid = false
				}
			case "count", "isempty", "top", "bottom":
			default:
				valid = false
			}
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return valid
}
