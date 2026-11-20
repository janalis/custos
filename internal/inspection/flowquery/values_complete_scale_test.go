package flowquery

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// TestPossibleValuesCompletePropertiesLinear guards the per-class index of
// property writes: resolving `$this->p` once per method of a class with n
// methods must stay linear in n (it rescanned the whole class per lookup,
// 4.8 s on Moodle's 900 KB tcpdf.php).
func TestPossibleValuesCompletePropertiesLinear(t *testing.T) {
	// Runs alternate between the sizes, so that both see the same load,
	// and the fastest run of each counts.
	small, large := propertyLookup(t, 300), propertyLookup(t, 1200)
	bestSmall, bestLarge := time.Duration(1<<63-1), time.Duration(1<<63-1)
	for range 7 {
		bestSmall, bestLarge = min(bestSmall, small()), min(bestLarge, large())
	}
	ratio := float64(bestLarge) / float64(bestSmall)
	t.Logf("300: %v, 1200: %v (x%.1f)", bestSmall, bestLarge, ratio)
	if ratio > 9 {
		t.Errorf("4x more methods took x%.1f longer (super-linear)", ratio)
	}
}

// propertyLookup returns a function timing one resolution of every
// `$this->p` probe in a class of n methods.
func propertyLookup(t *testing.T, n int) func() time.Duration {
	t.Helper()
	var b strings.Builder
	b.WriteString("<?php class C {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "private $p%d = 'd';\nfunction m%d() { $this->p%d = 'e'; $x = 1; $y = [$x, $x, $x]; probe($this->p%d); }\n", i, i, i, i)
	}
	b.WriteString("}\n")
	src := b.String()
	return func() time.Duration {
		f := Parse(t, src)
		var probes []syntax.Expr
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(c) == "probe" {
				args, _ := astquery.CallArgValues(c)
				probes = append(probes, args[0])
			}
			return true
		})
		runtime.GC()
		start := time.Now()
		for _, p := range probes {
			if vals, ok := PossibleValuesComplete(f, p); !ok || len(vals) != 2 {
				t.Fatalf("%s: %d values, complete=%v", Text(f, p), len(vals), ok)
			}
		}
		return time.Since(start)
	}
}
