package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestReachable(t *testing.T) {
	src := `<?php function g($c) {
		a1();
		if ($c) { return; }
		a2();
		if ($c) { return; } else { throw new E(); }
		a3();
	}
	function h() { while (true) { x(); } a4(); }
	function k() { while (true) { break; } a5(); switch (1) { case 1: return; default: return; } a6(); }
	function m() { try { return; } catch (E $e) { a7(); } a8(); }`
	f := parse(t, src)
	want := map[string]bool{"a1": true, "a2": true, "a3": false, "a4": false, "a5": true, "a6": false, "a7": true, "a8": true}
	got := map[string]bool{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name := CallLastName(c); len(name) == 2 && name[0] == 'a' {
				got[name] = Reachable(c, EnclosingFuncLike(c))
			}
		}
		return true
	})
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Reachable(%s) = %v, want %v", k, got[k], v)
		}
	}
}
