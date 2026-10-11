package compressiondecoderformatmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestLocalScan covers inputs the shared resolvers still answer with a value,
// so the rule's own scan of the scope decides whether to report.
func TestLocalScan(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"CompressionDecoderFormatMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"bytes captured by reference before any call", "<?php function f(string $x) { $y = gzencode($x); $wipe = function () use (&$y) { $y = ''; }; return gzuncompress($y); }", 0},
		{"copy of bytes captured by reference", "<?php function f(string $x) { $y = gzencode($x); $wipe = function () use (&$y) { $y = ''; }; $z = $y; return gzuncompress($z); }", 0},
		{"copy of bytes captured by value", "<?php function f(string $x) { $y = gzencode($x); $peek = function () use ($y) { return $y; }; $z = $y; return gzuncompress($z); }", 1},
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
