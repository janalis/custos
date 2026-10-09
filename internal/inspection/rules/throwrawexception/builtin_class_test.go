package throwrawexception

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestIsBuiltinClass(t *testing.T) {
	f := syntax.Parse("mine.php", []byte("<?php class Mine extends \\RuntimeException {}"), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !isBuiltinClass(ix.Class("RuntimeException", phpversion.PHP84), phpversion.PHP84) {
		t.Fatal("RuntimeException should be builtin")
	}
	if isBuiltinClass(ix.Class("Mine", phpversion.PHP84), phpversion.PHP84) {
		t.Fatal("Mine should not be builtin")
	}
	if isBuiltinClass(nil, phpversion.PHP84) {
		t.Fatal("nil is not builtin")
	}
}
