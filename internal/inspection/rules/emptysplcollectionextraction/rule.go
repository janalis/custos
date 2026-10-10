// Package emptysplcollectionextraction implements the native EmptySplCollectionExtraction inspection.
package emptysplcollectionextraction

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
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
	if !semanticquery.NativeMethod(ctx, c, class, name.Value) || semanticquery.NativeConstruction(ctx, c.Var, class) == nil {
		return
	}
	if !knownHistory(ctx, c) {
		return
	}
	size := 0
	for _, n := range semanticquery.NativePriorCalls(ctx, c, c.Var, "insert", "push", "unshift", "enqueue", "extract", "pop", "shift", "dequeue") {
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

func guardedNonempty(ctx *analysis.Context, c *syntax.MethodCall) bool {
	for p := c.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if branch, ok := p.(*syntax.If); ok {
			if branch.Body.Span().Contains(c.Span()) && nonemptyCondition(ctx, branch.Cond, c.Var, true) {
				return true
			}
			if branch.Else != nil && branch.Else.Body.Span().Contains(c.Span()) && nonemptyCondition(ctx, branch.Cond, c.Var, false) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok {
			prev, exists := astquery.PrevStmt(ctx.File, st)
			if guard, ok := prev.(*syntax.If); exists && ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && nonemptyCondition(ctx, guard.Cond, c.Var, false) {
				return true
			}
		}
	}
	return false
}

func nonemptyCondition(ctx *analysis.Context, e, receiver syntax.Expr, truth bool) bool {
	e = syntax.UnwrapParens(e)
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		return nonemptyCondition(ctx, u.Expr, receiver, !truth)
	}
	if call, ok := e.(*syntax.MethodCall); ok && semanticquery.NativeSameValue(ctx, call.Var, receiver) && (semanticquery.NativeMethod(ctx, call, "SplDoublyLinkedList", "isEmpty") || semanticquery.NativeMethod(ctx, call, "SplHeap", "isEmpty") || semanticquery.NativeMethod(ctx, call, "SplPriorityQueue", "isEmpty")) {
		return !truth
	}
	if b, ok := e.(*syntax.Binary); ok {
		call, ok := b.Left.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeSameValue(ctx, call.Var, receiver) || !(semanticquery.NativeMethod(ctx, call, "SplDoublyLinkedList", "count") || semanticquery.NativeMethod(ctx, call, "SplHeap", "count") || semanticquery.NativeMethod(ctx, call, "SplPriorityQueue", "count")) {
			return false
		}
		value, known := semanticquery.NativeInt(ctx, b.Right)
		if !known {
			return false
		}
		if !truth {
			return (b.Op.Kind == syntax.TIsIdentical || b.Op.Kind == syntax.TIsEqual) && value == 0
		}
		return (b.Op.Kind == syntax.TGreater && value >= 0) || (b.Op.Kind == syntax.TIsGreaterOrEqual && value >= 1) || ((b.Op.Kind == syntax.TIsNotIdentical || b.Op.Kind == syntax.TIsNotEqual) && value == 0)
	}
	return false
}

// Only an entirely modeled history can establish an empty collection.
func knownHistory(ctx *analysis.Context, at *syntax.MethodCall) bool {
	valid, budget := true, 256
	scope := syntax.EnclosingVariableScope(at)
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			valid = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			valid = false
			return false
		}
		var args *syntax.ArgList
		switch call := n.(type) {
		case *syntax.FuncCall:
			args = call.Args
		case *syntax.MethodCall:
			args = call.Args
		case *syntax.StaticCall:
			args = call.Args
		}
		if args != nil {
			syntax.Inspect(args, func(child syntax.Node) bool {
				if variable, ok := child.(*syntax.Variable); ok && semanticquery.NativeSameValue(ctx, variable, at.Var) {
					valid = false
				}
				return true
			})
		}
		switch n := n.(type) {
		case *syntax.Assign:
			if n.ByRef {
				valid = false
			}
			if slot, ok := n.Var.(*syntax.ArrayDimFetch); ok && semanticquery.NativeSameValue(ctx, slot.Var, at.Var) {
				valid = false
			}
		case *syntax.MethodCall:
			if !semanticquery.NativeSameValue(ctx, n.Var, at.Var) {
				return true
			}
			name := ""
			if id, ok := n.Name.(*syntax.Identifier); ok {
				name = strings.ToLower(id.Value)
			}
			switch name {
			case "insert", "push", "unshift", "enqueue", "extract", "pop", "shift", "dequeue":
				if !semanticquery.NativeDominates(n, at) {
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
