package util

import (
	"testing"

	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestIsBuiltinClass(t *testing.T) {
	f := syntax.Parse("mine.php", []byte("<?php class Mine extends \\RuntimeException {}"), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !IsBuiltinClass(ix.Class("RuntimeException", phpver.PHP84), phpver.PHP84) {
		t.Fatal("RuntimeException should be builtin")
	}
	if IsBuiltinClass(ix.Class("Mine", phpver.PHP84), phpver.PHP84) {
		t.Fatal("Mine should not be builtin")
	}
	if IsBuiltinClass(nil, phpver.PHP84) {
		t.Fatal("nil is not builtin")
	}
}
