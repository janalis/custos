package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

// An integer literal Go cannot read (an invalid octal PHP 5 accepts) types
// as a plain int, without a value.
func TestUnreadableIntegerLiteral(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php $a = 09;`), syntax.Options{Version: phpversion.PHP56})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	e := NewEnv(f, names.New(f), index.New(nil), phpversion.PHP56, Options{})
	found := false
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if l, ok := n.(*syntax.Literal); ok {
			found = true
			if got := e.TypeOf(l).ShapeString(); got != "int" {
				t.Errorf("got %s want int", got)
			}
		}
		return true
	})
	if !found {
		t.Fatal("no literal")
	}
}
