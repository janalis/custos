package zipentryindexpastend

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestLocalScan covers inputs the shared resolvers still answer with a value,
// so the rule's own scan of the scope decides whether to report.
func TestLocalScan(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"ZipEntryIndexPastEnd"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"index captured by reference before any call", "<?php function f(ZipArchive $z) { $i = -1; $bump = function () use (&$i) { $i = 0; }; echo $z->getNameIndex($i); }", 0},
		{"index captured by reference inside an arrow function", "<?php function f(ZipArchive $z) { return fn() => [$i = -1, $bump = function () use (&$i) { $i = 0; }, $z->getNameIndex($i)]; }", 0},
		{"index captured by value inside an arrow function", "<?php function f(ZipArchive $z) { return fn() => [$i = -1, $peek = function () use ($i) { return $i; }, $z->getNameIndex($i)]; }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("scan.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
