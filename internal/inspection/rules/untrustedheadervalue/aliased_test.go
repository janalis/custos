package untrustedheadervalue

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A variable's reaching assignment proves nothing once the scope can rebind
// the variable behind it; the flow verdict is then kept.
func TestReviewAliasedOrigin(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"UntrustedHeaderValue"}})
	if err != nil {
		t.Fatal(err)
	}
	const (
		origin = "$n = $_GET['a'] + 1; "
		sink   = "header('X-N: ' . $n); "
	)
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"plain", origin + sink, 0},
		{"value capture", origin + "$f = function () use ($n) { return $n; }; " + sink, 0},
		{"other catch", origin + "try { run(); } catch (Exception $e) { } catch (Error) { } " + sink, 0},
		{"other reference", origin + sink + "$m = &$o->p;", 0},
		{"variable variable read", origin + sink + "echo $$k;", 0},
		{"reference capture", origin + "$f = function () use (&$n) { $n = $_GET['b']; }; $f(); " + sink, 0},
		{"capture beside a call", origin + "$x = (function () use (&$n) { $n = $_GET['b']; })() . strlen('a'); " + sink, 0},
		{"extract", origin + "extract($_GET); " + sink, 1},
		{"catch binding", origin + "try { run(); } catch (Exception $n) { } " + sink, 0},
		{"reference alias", origin + sink + "$m = &$n;", 1},
		{"element alias", origin + sink + "$m = &$n['k'];", 1},
		{"reference loop", origin + sink + "foreach ($n as &$m) { }", 1},
		{"reference argument", origin + sink + "settype($n, 'string');", 1},
		{"variable variable", origin + sink + "$$k = 1;", 1},
		{"globals table", origin + "$all = $GLOBALS; " + sink, 1},
		{"static declaration", "function probe() { " + origin + sink + "static $n; }", 1},
		{"global declaration", "function probe() { global $n; " + origin + sink + "}", 1},
		{"unrelated scopes", "function other() { extract($_POST); static $n; global $m; } " + origin + sink, 0},
		{"global written elsewhere", "function swap() { global $n; $n = $_GET['b']; } " + origin + "swap(); " + sink, 0},
		{"globals entry written elsewhere", "function swap() { $GLOBALS['n'] = $_GET['b']; } " + origin + "swap(); " + sink, 0},
		{"other globals entry", "function swap() { $GLOBALS['m'] = $_GET['b']; } " + origin + "swap(); " + sink, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sources := []string{"<?php " + tc.body}
			if !strings.HasPrefix(tc.body, "function ") {
				sources = append(sources, "<?php function probe() { "+tc.body+" }")
			}
			for _, source := range sources {
				f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe: %v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("%s\nfindings: got %v, want %d", source, got, tc.findings)
				}
			}
		})
	}
}
