package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// Before PHP 8.0 array_key_exists() also searched an object's properties: a
// true call only rules out the values it fails on.
func TestKeyExistsBeforePHP8(t *testing.T) {
	src := `<?php
/** @param array|object|false|null $a */
function f($a, $k) { if (array_key_exists($k, $a)) { u($a); } }`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP74})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP74)
	var got string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
				got = env.TypeOf(c.Args.Args[0].(*syntax.Arg).Value).String()
			}
		}
		return true
	})
	if got != "array|object" {
		t.Fatalf("got %s", got)
	}
}
