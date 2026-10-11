package collatorsortdiscardsrequiredkeys

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const head = `<?php function f(Collator $c) { $m = ["k" => "b", "j" => "a"]; `
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"reference capture", `$g = function () use (&$m) { $m = ["z", "y"]; }; $g(); $c->sort($m); }`, 0},
		{"static closure reference capture", `$g = static function () use (&$m) { $m = []; }; $c->sort($m); }`, 0},
		{"copy of captured local", `$g = function () use (&$m) { $m = []; }; $n = $m; $c->sort($n); }`, 0},
		{"self copy", `$g = function () use (&$m) { $m = []; }; $n = $m; $n = $n; $c->sort($n); }`, 0},
		{"catch binding", `try { g(); } catch (Exception $m) { } $c->sort($m); }`, 0},
		{"value capture", `$g = function () use ($m) { return $m; }; $c->sort($m); }`, 1},
		{"unrelated reference capture", `$g = function () use (&$o) { $o = 1; }; $c->sort($m); }`, 1},
		{"capture in nested scope", `$g = function () { return function () use (&$m) { $m = []; }; }; $c->sort($m); }`, 1},
		{"capture of arrow function copy", `$g = fn () => function () use (&$m) { $m = []; }; $c->sort($m); }`, 1},
		{"catch binding of another variable", `try { g(); } catch (Exception $e) { } $c->sort($m); }`, 1},
		{"untyped catch", `try { g(); } catch (Exception) { } $c->sort($m); }`, 1},
		{"literal argument", `$c->sort(["k" => "b"]); }`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(head + tc.body)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if tc.findings == 0 && string(res.Source) != string(src) {
				t.Fatalf("unexpected fix: %s", res.Source)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
