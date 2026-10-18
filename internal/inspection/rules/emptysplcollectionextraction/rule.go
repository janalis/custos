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

// guardedNonempty recognizes a size test on the receiver that controls whether
// the call runs: a branch, a conditional operand or a loop condition.
func guardedNonempty(ctx *analysis.Context, c *syntax.MethodCall) bool {
	for child, p := syntax.Node(c), c.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch branch := p.(type) {
		case *syntax.If:
			if child == branch.Cond {
				break
			}
			if child == branch.Body {
				if nonemptyCondition(ctx, branch.Cond, c.Var, true) {
					return true
				}
				break
			}
			// An elseif or else body runs only after every earlier condition failed.
			if nonemptyCondition(ctx, branch.Cond, c.Var, false) {
				return true
			}
			for _, alt := range branch.ElseIfs {
				if child == alt {
					break
				}
				if nonemptyCondition(ctx, alt.Cond, c.Var, false) {
					return true
				}
			}
		case *syntax.ElseIf:
			if child == branch.Body && nonemptyCondition(ctx, branch.Cond, c.Var, true) {
				return true
			}
		case *syntax.Ternary:
			if child != branch.Cond && nonemptyCondition(ctx, branch.Cond, c.Var, child == branch.Then) {
				return true
			}
		case *syntax.Binary:
			if child != branch.Right {
				break
			}
			switch branch.Op.Kind {
			case syntax.TBooleanAnd, syntax.TAnd:
				if nonemptyCondition(ctx, branch.Left, c.Var, true) {
					return true
				}
			case syntax.TBooleanOr, syntax.TOr:
				if nonemptyCondition(ctx, branch.Left, c.Var, false) {
					return true
				}
			}
		case *syntax.While:
			if child == branch.Body && nonemptyCondition(ctx, branch.Cond, c.Var, true) {
				return true
			}
		case *syntax.For:
			if child == branch.Body && len(branch.Cond) > 0 && nonemptyCondition(ctx, branch.Cond[len(branch.Cond)-1], c.Var, true) {
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
		// A true conjunction proves both operands; a false disjunction refutes both.
		switch b.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (nonemptyCondition(ctx, b.Left, receiver, true) || nonemptyCondition(ctx, b.Right, receiver, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (nonemptyCondition(ctx, b.Left, receiver, false) || nonemptyCondition(ctx, b.Right, receiver, false))
		}
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
