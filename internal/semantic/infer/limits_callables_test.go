package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// The inference added for closures, out parameters, conditional return
// types and property writes stays linear on hostile bodies (one construct
// repeated 20k times, deep nesting, conditionals at the doc-type caps).
func TestCallableInferenceBounded(t *testing.T) {
	const n = 20000
	method := func(line string) string {
		return "<?php\nclass X { public function m(): int { return 1; } }\nfinal class C {\n private ?X $p = null;\n public function f(string $s, $u) {\n $n = null; $f0 = fn() => 1;\n" +
			strings.Repeat(line+"\n", n) + " }\n}\n"
	}
	var chain strings.Builder
	chain.WriteString("<?php\nfunction f() {\n $f0 = fn() => 1;\n")
	for i := 1; i < n; i++ {
		fmt.Fprintf(&chain, " $f%d = fn() => $f%d();\n", i, i-1)
	}
	fmt.Fprintf(&chain, " $y = $f%d();\n}\n", n-1)
	// A conditional nested to the doc-type length cap, called n times.
	cond := "int"
	for len(cond) < 4000 {
		cond = "($x is string ? float : " + cond + ")"
	}
	condSrc := "<?php\n/** @return " + cond + " */\nfunction g($x) {}\nfunction f(string $s) {\n" + strings.Repeat(" $y = g($s);\n", n) + "}\n"
	cases := map[string]string{
		"property write and read": method(" $this->p = new X; $y = $this->p;"),
		"builtin calls after a write": "<?php\nclass X {}\nfinal class C {\n private ?X $p = null;\n public function f() {\n $this->p = new X;\n" +
			strings.Repeat(" strlen('a'); $y = $this->p;\n", n) + " }\n}\n",
		"out parameters":     method(" preg_match('/a/', $s, $m); $y = $m;"),
		"param-out":          method(" $this->out($v); $y = $v;"),
		"writes into null":   method(" $n['k'] = 1; $y = $n;"),
		"closure calls":      method(" $g = fn() => new X; $y = $g();"),
		"array_map":          method(" $y = array_map(fn($v) => $v . 'a', [$s]);"),
		"closure chain":      chain.String(),
		"conditional return": condSrc,
		"nested arrows":      "<?php\nfunction f() {\n $y = " + strings.Repeat("fn() => ", 1000) + "1;\n $z = $y();\n}\n",
		"nested closures": "<?php\nfunction f() {\n $y = " + strings.Repeat("function () { return ", 1000) + "1;" +
			strings.Repeat(" };", 1000) + "\n $z = $y();\n}\n",
	}
	for name, src := range cases {
		d := typeAll(t, src)
		t.Logf("%s: %v", name, d)
		if d > testbudget.Of(2*time.Second) && !raceEnabled {
			t.Errorf("%s: took %v", name, d)
		}
	}
}
