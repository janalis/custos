package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

// NarrowMember narrows an assumed type at member and element reads by
// their guards, whatever their declarations say.
func TestNarrowMember(t *testing.T) {
	src := `<?php
function g(): ?string { return null; }
class A {
    public $p;
    public static $s;
    function f($v, $flag) {
        $this->p = g();
        if ($this->p !== null) { u($this->p); }
        if ($this->p === null) { return; }
        u($this->p);
        $this->p = g();
        u($this->p);
        if (self::$s) { u(self::$s); }
        if (isset($v['k'])) { u($v['k']); }
        if ($v !== null) { u($v); }
        u(g());
        if ($this->p !== null) { $this->p = g(); u($this->p); }
    }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
				got = append(got, env.NarrowMember(c.Args.Args[0].(*syntax.Arg).Value, types.Of("string", "null")).String())
			}
		}
		return true
	})
	want := []string{"string", "string", "null|string", "string", "string", "null|string", "null|string", "null|string"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("read %d: got %s want %s", i, got[i], want[i])
		}
	}
}
