package nonvoidfunctionfallsthrough

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestMissingExpression(t *testing.T) {
	if mayCall(nil) {
		t.Fatal("missing expression cannot contain a call")
	}
}

func TestMayCallSiblings(t *testing.T) {
	for src, want := range map[string]bool{
		"f() && $a + 1;":   true,
		"$a / 2 && $b;":    true,
		"$a + 1 && $b;":    false,
		"eval($code);":     true,
		"clone $object;":   true,
		"`ls` . $a + 1;":   true,
		"\"$a\" . $b + 1;": false,
	} {
		f := syntax.Parse("case.php", []byte("<?php "+src), syntax.Options{})
		stmt := f.Stmts[0].(*syntax.ExprStmt)
		if got := mayCall(stmt.Expr); got != want {
			t.Errorf("%s: got %v, want %v", src, got, want)
		}
	}
}
