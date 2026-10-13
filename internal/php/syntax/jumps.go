package syntax

import (
	"slices"
	"sync/atomic"
)

// listJumps caches, on the node that owns a statement list, what the
// questions about `goto` ask of it again and again: the Tail of the list
// and, for the body of a variable scope, the jumps of the scope.
type listJumps struct {
	tail  atomic.Pointer[Tail]
	gotos atomic.Pointer[[]*Goto]
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
	body, ok := VariableScopeBody(scope).(*Block)
	if !ok {
		return nil
	}
	if all := body.jumps.gotos.Load(); all != nil {
		return *all
	}
	all := []*Goto{}
	inspectVariableScope(body, func(n Node) {
		if jump, ok := n.(*Goto); ok {
			all = append(all, jump)
		}
	})
	body.jumps.gotos.Store(&all)
	return all
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
