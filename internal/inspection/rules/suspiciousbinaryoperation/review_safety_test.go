package suspiciousbinaryoperation

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
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		// D4: the replacement keeps its grouping under tighter operators.
		{
			"negated under cast", "<?php function f(?int $n) { return (int) !($n < 10); }", 1,
			"<?php function f(?int $n) { return (int) ($n >= 10); }",
		},
		{
			"negated under not", "<?php function f(?int $n) { return !!($n < 10); }", 1,
			"<?php function f(?int $n) { return !($n >= 10); }",
		},
		{
			"negated under minus", "<?php function f(?int $n) { return -!($n <= 10); }", 1,
			"<?php function f(?int $n) { return -($n > 10); }",
		},
		{
			"negated compared", "<?php function f(?int $n, $x) { return !($n < 10) > $x; }", 2,
			"<?php function f(?int $n, $x) { return ($n >= 10) > $x; }",
		},
		{
			"negated equal", "<?php function f(?int $n, $x) { return !($n < 10) == $x; }", 1,
			"<?php function f(?int $n, $x) { return ($n >= 10) == $x; }",
		},
		{
			"negated in logical", "<?php function f(?int $n, $x) { return !($n < 10) && $x; }", 1,
			"<?php function f(?int $n, $x) { return $n >= 10 && $x; }",
		},
		{
			"negated in ternary", "<?php function f(?int $n) { return !($n < 10) ? 1 : 2; }", 1,
			"<?php function f(?int $n) { return $n >= 10 ? 1 : 2; }",
		},
		{
			"negated open right", "<?php function f(?int $n, int $m, $x) { return !($n < $m = 5) && $x; }", 1,
			"<?php function f(?int $n, int $m, $x) { return ($n >= $m = 5) && $x; }",
		},
		// D4: no fix unless both operands are always ordered.
		{"negated float", "<?php function f(?float $v) { return !($v < 10); }", 1, ""},
		{"negated array", "<?php function f(?array $x, array $y) { return !($x < $y); }", 1, ""},
		{"negated string", "<?php function f(?string $s) { return !($s < 'b'); }", 1, ""},
		{"negated float right", "<?php function f(?int $l, float $g) { return !($l <= $g); }", 1, ""},
		{"negated unknown right", "<?php function f(?int $l, $g) { return !($l <= $g); }", 1, ""},
		{"negated constant right", "<?php function f(?int $l) { return !($l < NAN); }", 1, ""},
		{
			"negated float literal", "<?php function f(?int $l) { return !($l < 1.5); }", 1,
			"<?php function f(?int $l) { return $l >= 1.5; }",
		},
		{
			"negated typed right", "<?php function f(?int $l, ?bool $m) { return !($l <= $m); }", 1,
			"<?php function f(?int $l, ?bool $m) { return $l > $m; }",
		},
		// D6: directly under `!` the moved comparison keeps a group.
		{
			"misplaced under not", "<?php function f(string $s) { if (!strlen($s >= 2)) {} }", 1,
			"<?php function f(string $s) { if (!(strlen($s) >= 2)) {} }",
		},
		{
			"misplaced under grouped not", "<?php function f(string $s) { if (!(strlen($s >= 2))) {} }", 1,
			"<?php function f(string $s) { if (!(strlen($s) >= 2)) {} }",
		},
		{
			"misplaced open right", "<?php function f(string $s, int $m, $x) { if (strlen($s >= $m = 2) && $x) {} }", 1,
			"<?php function f(string $s, int $m, $x) { if ((strlen($s) >= $m = 2) && $x) {} }",
		},
		// D5: operands whose two evaluations may differ.
		{"same yield", "<?php function g() { $x = (yield) === (yield); }", 0, ""},
		{"same yield from", "<?php function g($it) { $x = (yield from $it) === (yield from $it); }", 0, ""},
		{"same include", "<?php $x = (include 'x.php') == (include 'x.php');", 0, ""},
		{"same eval", "<?php $x = eval('return 1;') == eval('return 1;');", 0, ""},
		{"same backtick", "<?php $x = `date` == `date`;", 0, ""},
		{"same readline", "<?php $x = readline() === readline();", 0, ""},
		{"same stream", "<?php $x = stream_get_contents($h) == stream_get_contents($h);", 0, ""},
		{"same fetch", "<?php $x = mysqli_fetch_row($r) == mysqli_fetch_row($r);", 0, ""},
		{"same namespaced", "<?php namespace App; $x = \\readline() === \\readline();", 0, ""},
		{"same string", "<?php $x = \"a$b\" == \"a$b\";", 1, ""},
		{"same pure call", "<?php $x = strlen($s) == strlen($s);", 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
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
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
