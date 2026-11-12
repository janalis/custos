package semanticquery

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestRunsBefore covers runsBefore on the accesses named first and second
// (a method call, a property read or a static call) of each snippet.
func TestRunsBefore(t *testing.T) {
	cases := []struct {
		name, code string
		want       bool
	}{
		{"statements", `$x->first(); $x->second();`, true},
		{"nested block", `{ $x->first(); } $x->second();`, true},
		{"if condition", `if ($x->first()) { $x->second(); }`, true},
		{"if body", `if ($a) { $x->first(); } $x->second();`, false},
		{"elseif condition", `if ($a) {} elseif ($x->first()) { $x->second(); }`, true},
		{"while condition", `while ($x->first()) { $x->second(); }`, true},
		{"foreach subject", `foreach ($x->first() as $v) { $x->second(); }`, true},
		{"for init", `for ($i = $x->first(); $i < 1; $i++) { $x->second(); }`, true},
		{"after a while", `while ($x->first()) {} $x->second();`, true},
		{"after a foreach", `foreach ($x->first() as $v) {} $x->second();`, true},
		{"after a switch", `switch ($x->first()) {} $x->second();`, true},
		{"after a for", `for ($x->first();;) {} $x->second();`, true},
		{"for step", `for (;; $x->first()) {} $x->second();`, false},
		{"for condition", `for (; $x->first();) { $x->second(); }`, false},
		{"switch subject", `switch ($x->first()) { case 1: $x->second(); }`, true},
		{"same case", `switch ($a) { case 1: $x->first(); $x->second(); }`, true},
		{"other case", `switch ($a) { case 1: $x->first(); break; case 2: $x->second(); }`, false},
		{"match subject", `echo match ($x->first()) { 1 => $x->second() };`, true},
		{"match arm", `echo match ($a) { 1 => $x->first() }; $x->second();`, false},
		{"ternary branch", `echo $a ? $x->first() : 1; $x->second();`, false},
		{"ternary condition", `echo $x->first() ? $x->second() : 1;`, true},
		{"operands", `echo $x->first() + $x->second();`, true},
		{"right operand", `echo 'a' . $x->first(); $x->second();`, true},
		{"short-circuit", `$a && $x->first(); $x->second();`, false},
		{"coalescing assignment", `$a ??= $x->first(); $x->second();`, false},
		{"assignment", `$a = $x->first(); $x->second();`, true},
		{"within an assignment", `$x->first()->p = $x->second();`, false},
		{"arguments", `f($x->first(), $x->second());`, true},
		{"receiver then arguments", `$x->first()->m($x->second());`, true},
		{"class then arguments", `$x->first()::m($x->second());`, true},
		{"new class then arguments", `new ($x->first())($x->second());`, true},
		{"callee then arguments", `($x->first())($x->second());`, true},
		{"array then key", `echo $x->first()[$x->second()];`, true},
		{"object then name", `echo $x->first()->{$x->second()};`, true},
		{"after a nullsafe call", `$y?->m($x->first()); $x->second();`, false},
		{"key after a nullsafe read", `echo $y?->p[$x->first()]; $x->second();`, false},
		{"name after a nullsafe read", `echo $y?->p->{$x->first()}; $x->second();`, false},
		{"isset", `echo isset($a, $x->first); $x->second();`, false},
		{"try", `try { $x->first(); } finally {} $x->second();`, false},
		{"do-while body", `do { $x->first(); } while ($x->second());`, false},
		{"returned items", `return [$x->first(), $x->second()];`, true},
		{"echo list", `echo $x->first(), $x->second();`, true},
		{"key then value", `return [$x->first() => $x->second()];`, true},
		{"unary", `echo -$x->first(); $x->second();`, true},
		{"arrow function", `$f = fn() => $x->first(); $x->second();`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			first, second := markedAccesses(t, "<?php function f($x, $y, $a) { "+c.code+" }")
			if got := runsBefore(first, second); got != c.want {
				t.Errorf("runsBefore = %v, want %v", got, c.want)
			}
		})
	}
	for _, src := range []string{"<?php $x->first(); $x->second();", "<?php namespace A; $x->first(); $x->second();"} {
		if first, second := markedAccesses(t, src); !runsBefore(first, second) {
			t.Errorf("%s: top-level statements run in order", src)
		}
	}
}

// markedAccesses returns the accesses of src named first and second.
func markedAccesses(t *testing.T, src string) (first, second syntax.Node) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		var name syntax.Node
		switch a := n.(type) {
		case *syntax.MethodCall:
			name = a.Name
		case *syntax.PropertyFetch:
			name = a.Name
		}
		if id, ok := name.(*syntax.Identifier); ok {
			switch id.Value {
			case "first":
				first = n
			case "second":
				second = n
			}
		}
		return true
	})
	if first == nil || second == nil {
		t.Fatal("no first or second access")
	}
	return first, second
}
