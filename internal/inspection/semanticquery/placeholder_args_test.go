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

// TestNativePartial: a partial application of a function, a method or a
// static method makes a closure; a call or any other expression does not.
func TestNativePartial(t *testing.T) {
	for src, want := range map[string]bool{
		"<?php f(?);":      true,
		"<?php $o->m(?);":  true,
		"<?php C::m(...);": true,
		"<?php f(1);":      false,
		"<?php $o->m();":   false,
		"<?php C::m(1);":   false,
		"<?php $x;":        false,
	} {
		f := syntax.Parse("p.php", []byte(src), syntax.Options{Version: phpversion.PHP86})
		if got := nativePartial(f.Stmts[0].(*syntax.ExprStmt).Expr); got != want {
			t.Errorf("%s: %v, want %v", src, got, want)
		}
	}
}
