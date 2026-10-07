package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestStmtNeighbours(t *testing.T) {
	f := parse(t, "<?php a(); function g() { b(); /* c */ c(); } if ($x) d();")
	if _, ok := PrevStmt(f, f.Stmts[0]); ok {
		t.Fatal("first top-level stmt has no previous")
	}
	if n, ok := NextStmt(f, f.Stmts[0]); !ok || n != f.Stmts[1] {
		t.Fatal("NextStmt top level")
	}
	body := f.Stmts[1].(*syntax.Function).Body
	stmts, ok := BlockStmts(body)
	if !ok || len(stmts) != 2 {
		t.Fatalf("BlockStmts = %d, %v", len(stmts), ok)
	}
	if p, ok := PrevStmt(f, stmts[1]); !ok || text(f, p) != "b();" {
		t.Fatal("PrevStmt in block")
	}
	if _, ok := NextStmt(f, stmts[1]); ok {
		t.Fatal("last stmt has no next")
	}
	unbraced := f.Stmts[2].(*syntax.If).Body
	if _, _, ok := StmtList(f, unbraced); ok {
		t.Fatal("unbraced body is not in a list")
	}
	if _, ok := BlockStmts(unbraced); ok {
		t.Fatal("unbraced body is not a block")
	}
}

func TestStmtListForeignStmt(t *testing.T) {
	f := parse(t, "<?php a();")
	g := parse(t, "<?php b(); c();")
	if _, _, ok := StmtList(f, g.Stmts[1]); ok {
		t.Fatal("a statement of another file is not in f's list")
	}
}
