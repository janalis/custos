package util

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"custos/internal/syntax"
)

// TestPossibleValuesCompletePropertiesLinear guards the per-class index of
// property writes: resolving `$this->p` once per method of a class with n
// methods must stay linear in n (it rescanned the whole class per lookup,
// 4.8 s on Moodle's 900 KB tcpdf.php).
func TestPossibleValuesCompletePropertiesLinear(t *testing.T) {
	small, large := propertyLookupTime(t, 300), propertyLookupTime(t, 1200)
	ratio := float64(large) / float64(small)
	t.Logf("300: %v, 1200: %v (x%.1f)", small, large, ratio)
	if ratio > 9 {
		t.Errorf("4x more methods took x%.1f longer (super-linear)", ratio)
	}
}

func propertyLookupTime(t *testing.T, n int) time.Duration {
	t.Helper()
	var b strings.Builder
	b.WriteString("<?php class C {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "private $p%d = 'd';\nfunction m%d() { $this->p%d = 'e'; $x = 1; $y = [$x, $x, $x]; probe($this->p%d); }\n", i, i, i, i)
	}
	b.WriteString("}\n")
	best := time.Duration(1<<63 - 1)
	for run := 0; run < 5; run++ {
		f := parse(t, b.String())
		var probes []syntax.Expr
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && CallLastName(c) == "probe" {
				args, _ := CallArgValues(c)
				probes = append(probes, args[0])
			}
			return true
		})
		runtime.GC()
		start := time.Now()
		for _, p := range probes {
			if vals, ok := PossibleValuesComplete(f, p); !ok || len(vals) != 2 {
				t.Fatalf("%s: %d values, complete=%v", text(f, p), len(vals), ok)
			}
		}
		best = min(best, time.Since(start))
	}
	return best
}
