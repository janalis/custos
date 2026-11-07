package syntax

import "testing"

func TestNodeIdentityUsesStorageInsteadOfSourceRange(t *testing.T) {
	first, second := &Literal{}, &Literal{}
	first.span = Span{Start: 5, End: 9}
	second.span = first.span
	firstKey := Identity(first)
	if firstKey == Identity(second) {
		t.Fatal("distinct nodes with equal spans collided")
	}
	first.span.End++
	if firstKey != Identity(first) {
		t.Fatal("changing a range changed node identity")
	}
	if Identity(nil) != (NodeIdentity{}) || Identity(first) == Identity(nil) {
		t.Fatal("nil identity collided with a node")
	}
	other := &Variable{}
	other.span = first.span
	nodes := map[NodeIdentity]int{Identity(first): 1, Identity(second): 2, Identity(other): 3}
	if len(nodes) != 3 || nodes[Identity(first)] != 1 {
		t.Fatal("identity lookup lost a node")
	}
	left := Parse("same.php", []byte("<?php 1;"), Options{})
	right := Parse("same.php", []byte("<?php 1;"), Options{})
	if Identity(left.Stmts[0]) == Identity(right.Stmts[0]) {
		t.Fatal("identity survived a new parse")
	}
}

func BenchmarkNodeIdentityLookup(b *testing.B) {
	node := &Literal{}
	keys := map[NodeIdentity]bool{Identity(node): true}
	b.ReportAllocs()
	for b.Loop() {
		if !keys[Identity(node)] {
			b.Fatal("lost node")
		}
	}
}
