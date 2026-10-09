package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// narrowProbes are hostile inputs for condition narrowing: deep && / ||
// chains, alternating nestings (the negation of `A && B` and the truth of
// `A || B` narrow A both ways: 2^depth without the step budget), and
// thousands of elseifs, match arms and switch cases each read after the
// failed conditions before them.
func narrowProbes(n int) map[string]string {
	fn := func(body string) string {
		return "<?php\nfunction f(int|string|null|array $x, $c) {\n" + body + "\n}\n"
	}
	var chain, alt, elseifs, arms, armsTrue, cases strings.Builder
	depth := min(n, 1500) // syntax.MaxDepth bounds the nesting of one chain
	chain.WriteString("if ($x !== null")
	for i := 0; i < depth; i++ {
		chain.WriteString(" && $x !== null")
	}
	chain.WriteString(") { $y = $x; }\nif (!($x === null")
	for i := 0; i < depth; i++ {
		chain.WriteString(" || $x === null")
	}
	chain.WriteString(")) { return; }\n$z = $x;")
	e := "is_int($x)"
	for i := 0; i < 40+n/100; i++ {
		if i%2 == 0 {
			e = "(" + e + " && is_string($x))"
		} else {
			e = "(" + e + " || $x === null)"
		}
	}
	fmt.Fprintf(&alt, "if (!%s) { return; }\n", e)
	for i := 0; i < n; i++ {
		alt.WriteString("$y = $x;\n")
	}
	elseifs.WriteString("if ($x === 0) { $y = $x; }")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&elseifs, " elseif ($x === %d) { $y = $x; }", i)
	}
	elseifs.WriteString(" else { $y = $x; }")
	arms.WriteString("$y = match ($x) {\n")
	armsTrue.WriteString("$y = match (true) {\n")
	cases.WriteString("switch (true) {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&arms, "  %d, null => $x,\n", i)
		fmt.Fprintf(&armsTrue, "  is_int($x) && $c === %d => $x,\n", i)
		fmt.Fprintf(&cases, "  case $x === %d:\n  case is_string($x):\n    $y = $x;\n    break;\n", i)
	}
	arms.WriteString("  default => $x,\n};")
	armsTrue.WriteString("  default => $x,\n};")
	cases.WriteString("  default:\n    $y = $x;\n}")
	return map[string]string{
		"chains":      fn(chain.String()),
		"alternating": fn(alt.String()),
		"elseifs":     fn(elseifs.String()),
		"match":       fn(arms.String()),
		"match(true)": fn(armsTrue.String()),
		"switch":      fn(cases.String()),
	}
}

func TestNarrowingBounded(t *testing.T) {
	small, large := narrowProbes(2500), narrowProbes(10000)
	for name, src := range large {
		ds := typeAll(t, small[name])
		d := typeAll(t, src)
		t.Logf("%s: %v (n/4: %v)", name, d, ds)
		if d > testbudget.Of(2*time.Second) && !raceEnabled {
			t.Errorf("%s: took %v", name, d)
		}
	}
}
