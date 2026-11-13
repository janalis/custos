package project

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

func TestDependencyMetadataOwnsNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dependency.php")
	src := []byte("<?php namespace Sample; class Dependency { public function borrowedMethod(string $value): string { return $value; } }")
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatal(err)
	}
	original := extract
	t.Cleanup(func() { extract = original })
	var captured []byte
	extract = func(path string, source []byte, parse syntax.Options, options infer.Options) *index.FileSymbols {
		captured = source
		return original(path, source, parse, options)
	}
	owned := indexCached(path, syntax.Options{}, infer.Options{}, nil)
	borrowed := original(path, src, syntax.Options{}, infer.Options{})
	if !reflect.DeepEqual(owned, borrowed) {
		t.Fatal("ownership mode changed declarations")
	}
	at := bytes.Index(captured, []byte("borrowedMethod"))
	captured[at] = 'X'
	if owned.Classes[0].Methods["borrowedmethod"].Name != "borrowedMethod" {
		t.Fatal("dependency method retains source bytes")
	}
}
