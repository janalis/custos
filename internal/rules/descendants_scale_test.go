package rules

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// TestDescendantWalkLinear guards the per-class memo of descendant methods:
// the rules asking "does a subclass override this method?" walked every
// descendant once per method, so a base class with n methods and n indexed
// subclasses cost n² (SuiteCRM's vendored Google API client: Model.php, 10 KB,
// took 0.34 s next to its thousands of generated models).
func TestDescendantWalkLinear(t *testing.T) {
	small, large := descendantWalkTime(t, 300), descendantWalkTime(t, 1200)
	ratio := float64(large) / float64(small)
	t.Logf("300: %v, 1200: %v (x%.1f)", small, large, ratio)
	if ratio > 9 {
		t.Errorf("4x more methods and subclasses took x%.1f longer (super-linear)", ratio)
	}
}

func descendantWalkTime(t *testing.T, n int) time.Duration {
	t.Helper()
	var base, subs strings.Builder
	base.WriteString("<?php namespace N;\nclass Base {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&base, "public function m%d(array &$a) { self::h(); return 'x'; }\n", i)
	}
	base.WriteString("public function h() {}\n}\n")
	subs.WriteString("<?php namespace N;\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&subs, "class S%d extends Base { public function z%d() {} }\n", i, i)
	}
	opt := syntax.Options{Version: phpver.PHP84}
	bf := syntax.Parse("base.php", []byte(base.String()), opt)
	sf := syntax.Parse("subs.php", []byte(subs.String()), opt)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(bf))
	ix.Add(index.Extract(sf))
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpver.PHP84, Only: []string{
		"ReturnTypeCanBeDeclared", "DynamicInvocationViaScopeResolution", "ReferencingObjects",
	}})
	if err != nil {
		t.Fatal(err)
	}
	e.SetIndex(ix)
	best := time.Duration(1<<63 - 1)
	for run := 0; run < 5; run++ {
		runtime.GC()
		start := time.Now()
		if len(e.Analyze(bf)) == 0 {
			t.Fatal("no findings")
		}
		best = min(best, time.Since(start))
	}
	return best
}
