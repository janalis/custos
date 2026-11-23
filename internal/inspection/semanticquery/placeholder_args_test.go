package semanticquery

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// TestBindablePlaceholderArgs: a partial application binds no argument.
func TestBindablePlaceholderArgs(t *testing.T) {
	for _, src := range []string{"<?php f(...);", "<?php f(1, ?);", "<?php f(x: ?);"} {
		f := syntax.Parse("p.php", []byte(src), syntax.Options{Version: phpversion.PHP86})
		if got := bindableArgs(f.Stmts[0].(*syntax.ExprStmt).Expr); got != nil {
			t.Errorf("%s: %v", src, got)
		}
	}
}
