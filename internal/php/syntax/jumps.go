package syntax

import (
	"slices"
	"sync/atomic"
)

// listJumps caches, on the node that owns a statement list, what the
// questions about `goto` ask of it again and again: the Tail of the list
// and, for the body of a variable scope, the jumps and the labels of the
// scope and what a caller worked out from them (ScopeMemo).
type listJumps struct {
	tail  atomic.Pointer[Tail]
	scope atomic.Pointer[scopeJumps]
	memo  atomic.Value
}

// scopeJumps holds the `goto` statements and the labels of a variable scope.
type scopeJumps struct {
	gotos  []*Goto
	labels []*Label
}

// Tail describes the statements of a list from its first terminating one
// on: code that only a jump can reach.
type Tail struct {
	// Labels holds the labels written in those statements at any depth, in
	// source order, the ones of the functions, closures and classes declared
	// there left out.
	Labels []*Label
	first  int
	last   []int32
}

// LastTerminating returns the greatest index at or below i of a statement
// of the list that terminates. i must not lie before the first one.
func (t *Tail) LastTerminating(i int) int {
	return int(t.last[i-t.first])
}

// TailOf returns the Tail of the list owned by p (a *Block, *Case or
// *Namespace), or nil for another node or a list no statement ends. It is
// read once per list, so asking it for every node of a long list stays
// linear overall.
func TailOf(p Node) *Tail {
	var memo *listJumps
	var list []Stmt
	switch x := p.(type) {
	case *Block:
		memo, list = &x.jumps, x.Stmts
	case *Case:
		memo, list = &x.jumps, x.Stmts
	case *Namespace:
		memo, list = &x.jumps, x.Stmts
	default:
		return nil
	}
	if t := memo.tail.Load(); t != nil {
		return t
	}
	first := FirstTerminating(p)
	if first == len(list) {
		return nil
	}
	t := &Tail{first: first, last: make([]int32, len(list)-first)}
	last := first
	for i := first; i < len(list); i++ {
		if Terminates(list[i]) {
			last = i
		}
		t.last[i-first] = int32(last)
		inspectVariableScope(list[i], func(n Node) {
			if label, ok := n.(*Label); ok {
				t.Labels = append(t.Labels, label)
			}
		})
	}
	slices.SortStableFunc(t.Labels, func(a, b *Label) int { return int(a.Span().Start) - int(b.Span().Start) })
	memo.tail.Store(t)
	return t
}

// ScopeGotos returns the `goto` statements of a variable scope in source
// order, the ones of the functions, closures and classes declared in it left
// out. They are read once per scope. A scope without a block body has none.
func ScopeGotos(scope Node) []*Goto {
	if all := jumpsOf(scope); all != nil {
		return all.gotos
	}
	return nil
}

// ScopeLabels returns the labels of a variable scope in source order, read
// with its `goto` statements and with the same ones left out.
func ScopeLabels(scope Node) []*Label {
	if all := jumpsOf(scope); all != nil {
		return all.labels
	}
	return nil
}

// jumpsOf reads the jumps and the labels of a scope with a block body.
func jumpsOf(scope Node) *scopeJumps {
	body, ok := VariableScopeBody(scope).(*Block)
	if !ok {
		return nil
	}
	if all := body.jumps.scope.Load(); all != nil {
		return all
	}
	all := &scopeJumps{gotos: []*Goto{}}
	inspectVariableScope(body, func(n Node) {
		switch x := n.(type) {
		case *Goto:
			all.gotos = append(all.gotos, x)
		case *Label:
			all.labels = append(all.labels, x)
		}
	})
	slices.SortStableFunc(all.labels, func(a, b *Label) int { return int(a.Span().Start) - int(b.Span().Start) })
	body.jumps.scope.Store(all)
	return all
}

// ScopeMemo returns what build makes of a variable scope with a block body,
// kept on that body so that it is made once per scope; nil for another
// scope. One caller owns the slot: every call must ask for the same type.
func ScopeMemo[T any](scope Node, build func(scope Node) *T) *T {
	body, ok := VariableScopeBody(scope).(*Block)
	if !ok {
		return nil
	}
	if made, ok := body.jumps.memo.Load().(*T); ok {
		return made
	}
	made := build(scope)
	body.jumps.memo.Store(made)
	return made
}

// inspectVariableScope visits the nodes below n that share its variable
// scope: not the functions, closures and classes declared there.
func inspectVariableScope(n Node, visit func(Node)) {
	Inspect(n, func(child Node) bool {
		switch child.(type) {
		case *Function, *ClassLike, *Closure, *ArrowFunction:
			return false
		}
		visit(child)
		return true
	})
}
