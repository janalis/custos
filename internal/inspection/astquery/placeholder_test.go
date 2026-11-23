package astquery

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// TestPlaceholderArguments: a placeholder is no argument value, and an
// argument list holding one has no written arguments.
func TestPlaceholderArguments(t *testing.T) {
	f := syntax.Parse("p.php", []byte("<?php f(1, ?, ...);"), syntax.Options{Version: phpversion.PHP86})
	list := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args
	if ArgValue(list.Args[0]) == nil || ArgValue(list.Args[1]) != nil || ArgValue(list.Args[2]) != nil {
		t.Fatal("argument values")
	}
	if WrittenArguments(list) != nil {
		t.Fatal("a partial application has no written arguments")
	}
}
