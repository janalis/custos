package astquery

import (
	"testing"
)

func TestPrevStmtNoDoc(t *testing.T) {
	f := Parse(t, "<?php a(); // x\n b(); /** d */ c(); /* e */ d();")
	if p, ok := PrevStmtNoDoc(f, f.Stmts[1]); !ok || Text(f, p) != "a();" {
		t.Fatal("line comment must be skipped")
	}
	if _, ok := PrevStmtNoDoc(f, f.Stmts[2]); ok {
		t.Fatal("doc comment must stop")
	}
	if p, ok := PrevStmtNoDoc(f, f.Stmts[3]); !ok || Text(f, p) != "c();" {
		t.Fatal("block comment must be skipped")
	}
	if _, ok := PrevStmtNoDoc(f, f.Stmts[0]); ok {
		t.Fatal("first stmt has no previous")
	}
	if tok, ok := DocCommentBefore(f, f.Stmts[2].Span().Start); !ok || string(f.Src[tok.Start:tok.End]) != "/** d */" {
		t.Fatal("DocCommentBefore")
	}
	if _, ok := DocCommentBefore(f, f.Stmts[3].Span().Start); ok {
		t.Fatal("block comment is not a doc comment")
	}
}
