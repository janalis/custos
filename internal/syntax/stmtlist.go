package syntax

import (
	"sort"
	"sync/atomic"
)

// StmtListOf returns the statement list owned by p (a *Block, *Case or
// *Namespace); ok is false for any other node.
func StmtListOf(p Node) (list []Stmt, ok bool) {
	switch x := p.(type) {
	case *Block:
		return x.Stmts, true
	case *Case:
		return x.Stmts, true
	case *Namespace:
		return x.Stmts, true
	}
	return nil, false
}

// StmtIndex returns the index of s in list, or -1. Statement lists are in
// source order, so the lookup is a binary search on the start offset
// (O(log n)) with a linear fallback should the order assumption ever fail
// for a statement whose parent owns the list.
func StmtIndex(list []Stmt, s Node) int {
	if s == nil || len(list) == 0 {
		return -1
	}
	start := s.Span().Start
	i := sort.Search(len(list), func(i int) bool { return list[i].Span().Start >= start })
	for ; i < len(list) && list[i].Span().Start == start; i++ {
		if Node(list[i]) == s {
			return i
		}
	}
	if _, isStmt := s.(Stmt); !isStmt {
		return -1
	}
	for i, x := range list {
		if Node(x) == s {
			return i
		}
	}
	return -1
}

// FirstTerminating returns the index of the first statement of the list
// owned by p (*Block, *Case or *Namespace) that Terminates, or the list
// length when none does (also for other nodes: 0). The result is cached on
// the node, so callers asking "does a statement before index i terminate?"
// for many i stay linear overall instead of quadratic on long lists.
func FirstTerminating(p Node) int {
	var memo *int32
	var list []Stmt
	switch x := p.(type) {
	case *Block:
		memo, list = &x.term, x.Stmts
	case *Case:
		memo, list = &x.term, x.Stmts
	case *Namespace:
		memo, list = &x.term, x.Stmts
	default:
		return 0
	}
	// memo: 0 = not computed yet, otherwise index+1.
	if v := atomic.LoadInt32(memo); v > 0 {
		return int(v - 1)
	}
	i := 0
	for i < len(list) && !Terminates(list[i]) {
		i++
	}
	atomic.StoreInt32(memo, int32(i+1))
	return i
}
