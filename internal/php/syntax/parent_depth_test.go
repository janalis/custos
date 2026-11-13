package syntax

import "testing"

func TestLinkedDepth(t *testing.T) {
	for _, source := range []string{"<?php function f($x) { return $x + 1; }", "<?php if ($x) { echo $x; }", "<?php"} {
		f := Parse("test.php", []byte(source), Options{})
		InspectFile(f, func(n Node) bool {
			if n != f && n.Parent() == nil {
				t.Fatal("missing parent")
			}
			return true
		})
		if !linkParsedParents(f) {
			t.Fatal("ordinary tree rejected")
		}
	}
}

func BenchmarkLinkedParents(b *testing.B) {
	f := Parse("bench.php", []byte("<?php function f($x) { if ($x) { return $x + 1; } return 0; }"), Options{})
	b.ReportAllocs()
	for b.Loop() {
		linkParsedParents(f)
	}
}

func TestLinkedDepthBoundary(t *testing.T) {
	for _, depth := range []int{MaxDepth, MaxDepth + 1} {
		var stmt Stmt = &Block{}
		for i := 1; i < depth; i++ {
			stmt = &Block{Stmts: []Stmt{stmt}}
		}
		f := &File{Stmts: []Stmt{stmt}}
		if got := linkParsedParents(f); got != (TreeDepth(f.Stmts) <= MaxDepth) {
			t.Fatalf("depth %d: %v", depth, got)
		}
	}
}

func TestLinkedParentsEquivalent(t *testing.T) {
	for _, source := range []string{
		"<?php #[A] function f(?int $x): int { return $x ?? 1; }",
		"<?php namespace N; class C { function f($x) { return fn($y) => $x + $y; } }",
		"<?php if ($x) { echo ; }",
	} {
		f := Parse("test.php", []byte(source), Options{})
		type record struct {
			parent Node
			span   Span
		}
		original := map[Node]record{}
		InspectFile(f, func(n Node) bool { original[n] = record{n.Parent(), n.Span()}; return true })
		SetParents(f)
		InspectFile(f, func(n Node) bool {
			if got := (record{n.Parent(), n.Span()}); got != original[n] {
				t.Fatalf("%T differs: got %+v, want %+v", n, got, original[n])
			}
			return true
		})
	}
}

func TestParentSpanWidening(t *testing.T) {
	child := &Block{nodeBase: nodeBase{span: Span{Start: 2, End: 3}}}
	parent := &Block{nodeBase: nodeBase{span: Span{Start: 8, End: 9}}, Stmts: []Stmt{child}}
	f := &File{Stmts: []Stmt{parent}}
	SetParents(f)
	if parent.Span().Start != 2 || child.Parent() != parent {
		t.Fatal("child did not widen parent")
	}
	parent.span.Start = 8
	if !linkParsedParents(f) || parent.Span().Start != 2 {
		t.Fatal("linked child did not widen parent")
	}
}
