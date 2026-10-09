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

func TestClassRef(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php
namespace App;
class Base {}
class Foo extends Base {
    public function m() { self::a(); static::b(); parent::c(); Base::d(); $this::e(); }
}
`), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
	want := map[string]string{"a": `App\Foo`, "b": `App\Foo`, "c": `App\Base`, "d": `App\Base`, "e": `App\Foo`}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.StaticCall); ok {
			name := c.Name.(*syntax.Identifier).Value
			if got := env.ClassRef(c.Class); got != want[name] {
				t.Errorf("%s: got %q want %q", name, got, want[name])
			}
		}
		return true
	})
}
