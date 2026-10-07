package codestyle

import (
	"strings"
	"testing"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// benchSrc has no findings, so the benchmark measures the Check fast paths.
var benchSrc = "<?php\n" + strings.Repeat(`
class C {
    final public function a() { if (!$x && !($y)) { return !f(); } }
    private function b() { for (;;) ; while (g()) ; }
}
?><p><?= $t ?></p><?php
`, 200)

func BenchmarkCodeStyle(b *testing.B) {
	e, err := analysis.NewEngine(Rules(), analysis.Config{EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(benchSrc), syntax.Options{Version: phpver.Max})
	if len(f.Errors) > 0 {
		b.Fatal(f.Errors[0])
	}
	b.ReportAllocs()
	for b.Loop() {
		if fs := e.Analyze(f); len(fs) != 0 {
			b.Fatalf("unexpected findings: %v", fs)
		}
	}
}
