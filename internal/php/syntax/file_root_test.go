package syntax

import "testing"

func TestFileRoot(t *testing.T) {
	f := Parse("root.php", []byte("<?php return; dead: echo 1;"), Options{})
	if f.Parent() != nil || f.Kind() != KFile || f.Span().Len() != uint32(len(f.Src)) || IsVariableScope(f) {
		t.Fatal("file root contract")
	}
	if f.Stmts[0].Parent() != Node(f) {
		t.Fatal("top-level parent")
	}
	list, ok := StmtListOf(f)
	if !ok || len(list) != 3 || FirstTerminating(f) != 0 || FirstTerminating(f) != 0 || TailOf(f) == nil || len(TailOf(f).Labels) != 1 {
		t.Fatal("file statement cache")
	}
	visited := false
	InspectFile(f, func(n Node) bool { visited = visited || n == Node(f); return true })
	if visited {
		t.Fatal("InspectFile dispatched root")
	}
	empty := Parse("empty.php", nil, Options{})
	if FirstTerminating(empty) != 0 || TailOf(empty) != nil {
		t.Fatal("empty file tail")
	}
}
