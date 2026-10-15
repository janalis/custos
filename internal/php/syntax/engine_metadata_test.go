package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestModernSyntaxMinVersion(t *testing.T) {
	for _, tc := range []struct {
		source string
		kind   NodeKind
		want   phpversion.Version
	}{
		{`class C { const int X = 1; }`, KClassConst, phpversion.PHP83},
		{`class C { const X = 1; }`, KClassConst, phpversion.Min},
		{`class C { public int $p { get; } }`, KPropertyHook, phpversion.PHP84},
		{`class C { public private(set) int $p; }`, KProperty, phpversion.PHP84},
		{`class C { protected public(set) int $p; }`, KProperty, phpversion.PHP84},
		{`class C { public int $p; }`, KProperty, phpversion.Min},
		{`class C { function __construct(public protected(set) int $p) {} }`, KParam, phpversion.PHP84},
		{`function f(int $p) {}`, KParam, phpversion.Min},
		{`new A()->m();`, KNew, phpversion.PHP84},
		{`new A()();`, KNew, phpversion.PHP84},
		{`new A()->p;`, KNew, phpversion.PHP84},
		{`new A()[0];`, KNew, phpversion.PHP84},
		{`new A()::m();`, KNew, phpversion.PHP84},
		{`new A()::$p;`, KNew, phpversion.PHP84},
		{`new A()::X;`, KNew, phpversion.PHP84},
		{`new class {}->m();`, KNew, phpversion.PHP84},
		{`(new A())->m();`, KNew, phpversion.Min},
		{`new A();`, KNew, phpversion.Min},
		{`echo 1;`, KEcho, phpversion.Min},
	} {
		f := parse(t, "<?php "+tc.source, phpversion.PHP85)
		if len(f.Errors) != 0 {
			t.Fatalf("%s: %v", tc.source, f.Errors)
		}
		count := 0
		InspectFile(f, func(n Node) bool {
			if n.Kind() == tc.kind {
				count++
				if got := MinVersion(n); got != tc.want {
					t.Errorf("%s: %s want %s", tc.source, got, tc.want)
				}
			}
			return true
		})
		if count == 0 {
			t.Fatalf("missing target: %s", tc.source)
		}
	}
	if MinVersion(nil) != phpversion.Min {
		t.Fatal("nil")
	}
}
