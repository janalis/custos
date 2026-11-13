package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestReturnContract(t *testing.T) {
	env, f := envFor(t, `<?php
namespace N;
/** @template T */
class Box {
	/** @return list<T> */
	public function items(): array { return []; }
	/** @return $this */
	public function self(): static { return $this; }
	/** @return ($x is int ? string : int) */
	public function cond($x): int|string { return 1; }
	public function plain(): ?self { return null; }
}
trait Fluent {
	public function with(): static { return $this; }
}
`, nil)
	want := map[string][2]string{
		"items": {"array", "?unknown"},
		"self":  {`\N\Box`, `\N\Box`},
		"cond":  {"int|string", "?unknown"},
		"plain": {`\N\Box|null`, "?unknown"},
		"with":  {"?unknown", "?unknown"},
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if m, ok := n.(*syntax.Method); ok {
			c := env.ReturnContract(m)
			got := [2]string{c.Native.String(), c.Doc.String()}
			if got != want[m.Name.Value] {
				t.Errorf("%s: got %v, want %v", m.Name.Value, got, want[m.Name.Value])
			}
		}
		return true
	})
	if c := env.ReturnContract(f.Stmts[0]); !c.Native.IsUnknown() || !c.Doc.IsUnknown() {
		t.Errorf("non function-like: %+v", c)
	}
}
