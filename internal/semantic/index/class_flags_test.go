package index

import "testing"

func TestClassMixinAndConditionalFlags(t *testing.T) {
	fs := extract(t, "a.php", `<?php
namespace App {
    /** @mixin Builder */
    class Model {}
    /** @phpstan-mixin Builder */
    class Relation {}
    /** @psalm-mixin Builder */
    class Scope {}
    /** Plain doc. */
    class Plain {}
    {
        class Braced {}
    }
    if (!class_exists(Polyfill::class)) {
        class Polyfill {}
    }
    function make() {
        class Inner {}
        return new class {};
    }
}
declare(ticks=1) {
    class Declared {}
}
`)
	want := map[string][2]bool{ // FQN: {Mixin, Conditional}
		`App\Model`: {true, false}, `App\Relation`: {true, false}, `App\Scope`: {true, false},
		`App\Plain`: {false, false}, `App\Braced`: {false, false}, `App\Polyfill`: {false, true},
		`App\Inner`: {false, true}, `Declared`: {false, false},
	}
	seen := 0
	for _, c := range fs.Classes {
		if c.Anonymous {
			if c.Conditional {
				t.Errorf("anonymous class marked conditional")
			}
			continue
		}
		w, ok := want[c.FQN]
		if !ok {
			t.Fatalf("unexpected class %q", c.FQN)
		}
		seen++
		if c.Mixin != w[0] || c.Conditional != w[1] {
			t.Errorf("%s: mixin=%v conditional=%v, want %v", c.FQN, c.Mixin, c.Conditional, w)
		}
	}
	if seen != len(want) {
		t.Errorf("saw %d named classes, want %d", seen, len(want))
	}
}
