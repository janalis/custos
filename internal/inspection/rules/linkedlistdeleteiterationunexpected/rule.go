// Package linkedlistdeleteiterationunexpected implements the native LinkedListDeleteIterationUnexpected inspection.
package linkedlistdeleteiterationunexpected

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve linked-list elements during iteration."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LinkedListDeleteIterationUnexpected" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	f := n.(*syntax.Foreach)
	if _, ok := syntax.UnwrapParens(f.Expr).(*syntax.Clone); ok {
		return
	}
	classes := ctx.Types().Native().TypeOf(f.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	cl := ctx.Index().Class(classes[0], ctx.PHP)
	if cl == nil || !strings.HasPrefix(cl.File, "stubs/") || !ctx.Index().IsSubtype(cl.FQN, "SplDoublyLinkedList", ctx.PHP) {
		return
	}
	mode, last := int64(0), uint32(0)
	prior := map[syntax.Node]bool{}
	for _, n := range semanticquery.NativePriorCalls(ctx, f, f.Expr, "setIteratorMode") {
		c := n.(*syntax.MethodCall)
		known := false
		mode, known = semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "mode"))
		if !known {
			return
		}
		prior[n], last = true, max(last, n.Span().End)
	}
	if mode&1 != 0 && !modeUncertain(ctx, f, prior, last) {
		ctx.ReportNode(f.Expr, message)
	}
}

// modeUncertain reports whether the proven mode may have been replaced before
// the loop: by a mode change that is conditional, deferred or made through an
// unproven receiver, or through a closure or array holding the list.
func modeUncertain(ctx *analysis.Context, f *syntax.Foreach, prior map[syntax.Node]bool, last uint32) bool {
	name := ""
	if v, ok := syntax.UnwrapParens(f.Expr).(*syntax.Variable); ok {
		name = v.Name
	}
	scope := syntax.EnclosingVariableScope(f)
	// deferred finds a closure or array between n and the loop's scope; code in
	// a closure runs whenever it is invoked, wherever it is written.
	deferred := func(n syntax.Node, arrays bool) bool {
		for p := n.Parent(); p != nil && p != scope; p = p.Parent() {
			switch p.(type) {
			case *syntax.Closure, *syntax.ArrowFunction:
				return true
			case *syntax.Array:
				if arrays {
					return true
				}
			}
		}
		return false
	}
	uncertain := false
	visit := func(n syntax.Node) bool {
		if uncertain || n.Span().Start >= f.Span().Start {
			return false
		}
		switch n := n.(type) {
		case *syntax.MethodCall:
			id, ok := n.Name.(*syntax.Identifier)
			if !prior[n] && (!ok || strings.EqualFold(id.Value, "setIteratorMode")) && (n.Span().Start >= last || deferred(n, false)) {
				uncertain = true
			}
		case *syntax.Variable:
			if name != "" && n.Name == name && deferred(n, true) {
				uncertain = true
			}
		}
		return !uncertain
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return uncertain
}
