package splfixedarrayshrinkdiscardsvalues

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const filled = "$a = new SplFixedArray(4); $a[3] = 'tail'; "
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"plain", filled + "$a->setSize(2);", 1},
		{"method unset", filled + "$a->offsetUnset(3); $a->setSize(2);", 0},
		{"nullsafe method unset", filled + "$a?->offsetUnset(3); $a->setSize(2);", 0},
		{"method unset of a kept slot", filled + "$a->offsetUnset(1); $a->setSize(2);", 1},
		{"method unset then refill", filled + "$a->offsetUnset(3); $a[3] = 'again'; $a->setSize(2);", 1},
		{"method unset then method refill", filled + "$a->offsetUnset(3); $a->offsetSet(3, 'again'); $a->setSize(2);", 1},
		{"method unset of unknown slot", filled + "$a->offsetUnset($i); $a->setSize(2);", 0},
		{"method set", "$a = new SplFixedArray(4); $a->offsetSet(3, 'tail'); $a->setSize(2);", 1},
		{"method set of unknown slot", "$a = new SplFixedArray(4); $a->offsetSet($i, 'tail'); $a->setSize(2);", 0},
		{"conditional method set", "$a = new SplFixedArray(4); if ($c) { $a->offsetSet(3, 'tail'); } $a->setSize(2);", 0},
		{"refill after unset", filled + "unset($a[3]); $a[3] = 'again'; $a->setSize(2);", 1},
		{"conditional refill", filled + "unset($a[3]); if ($c) { $a[3] = 'again'; } $a->setSize(2);", 0},
		{"conditional unset", filled + "if ($c) { unset($a[3]); } $a->setSize(2);", 0},
		{"unset of unknown slot", filled + "unset($a[$i]); $a->setSize(2);", 0},
		{"append", filled + "try { $a[] = 1; } catch (Exception $e) {} $a->setSize(2);", 1},
		{"slot read", filled + "echo $a[3]; $a->setSize(2);", 1},
		{"used as offset", filled + "$other[$a] = 1; $a->setSize(2);", 0},
		{"capturing closure", filled + "$f = function () use ($a) { unset($a[3]); }; $f(); $a->setSize(2);", 0},
		{"unrelated closure", filled + "$f = function () use ($b) { $a = 1; }; $f(); $a->setSize(2);", 1},
		{"later closure", filled + "$a->setSize(2); $f = function () use ($a) {};", 1},
		{"capturing arrow function", filled + "$f = fn () => $a->offsetUnset(3); $f(); $a->setSize(2);", 0},
		{"unrelated arrow function", filled + "$f = fn ($b) => $b + 1; $a->setSize(2);", 1},
		{"alias", filled + "$b = $a; unset($b[3]); $a->setSize(2);", 0},
		{"reference", filled + "$b = &$a; $a->setSize(2);", 0},
		{"unknown function", filled + "release($a); $a->setSize(2);", 0},
		{"unknown method argument", filled + "$o->release($a); $a->setSize(2);", 0},
		{"count", filled + "echo count($a), sizeof($a); $a->setSize(2);", 1},
		{"inspecting builtins", filled + "var_dump($a); print_r($a); echo iterator_count($a); $a->setSize(2);", 1},
		{"clone", filled + "$b = clone $a; $a->setSize(2);", 1},
		{"spread argument", filled + "var_dump(...$a); $a->setSize(2);", 0},
		{"user function named like a builtin", "namespace N; function var_dump($x) { unset($x[3]); } " + filled + "var_dump($a); $a->setSize(2);", 0},
		{"foreach", filled + "foreach ($a as $v) { echo $v; } $a->setSize(2);", 1},
		{"foreach target", filled + "foreach ($list as $a) {} $a->setSize(2);", 0},
		{"reader methods", filled + "echo $a->getSize(), $a->offsetGet(3), count($a->toArray()); $a->setSize(2);", 1},
		{"unknown method", filled + "$a->__unserialize([]); $a->setSize(2);", 0},
		{"dynamic method", filled + "$a->$m(3); $a->setSize(2);", 0},
		{"used as method name", filled + "$o->$a(); $a->setSize(2);", 0},
		{"stored elsewhere", filled + "$list = [$a]; $a->setSize(2);", 0},
		// Only the first, proven shrink below is reported in these two.
		{"earlier shrink", filled + "$a->setSize(3); $a->setSize(4); $a->setSize(2);", 1},
		{"conditional shrink", filled + "if ($c) { $a->setSize(3); } $a->setSize(2);", 1},
		{"function scope", "function f() { " + filled + "$a->setSize(2); }", 1},
		{"function scope release", "function f() { " + filled + "$a->offsetUnset(3); $a->setSize(2); }", 0},
		{"other scope", "function g() { $a->offsetUnset(3); } " + filled + "$a->setSize(2);", 1},
		{"direct factory receiver", "SplFixedArray::fromArray(['x', 'y'])->setSize(1);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
