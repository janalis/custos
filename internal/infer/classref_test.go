package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestClassRef(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php
namespace App;
class Base {}
class Foo extends Base {
    public function m() { self::a(); static::b(); parent::c(); Base::d(); $this::e(); }
}
`), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
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
