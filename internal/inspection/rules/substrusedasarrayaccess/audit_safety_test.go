package substrusedasarrayaccess

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestAuditSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"negative runtime", "<?php $s = 'abc'; $offset = -4; substr($s, $offset, 1);", 0, ""},
		{"unknown runtime", "<?php function f(string $s, int $offset) { return substr($s, $offset, 1); }", 0, ""},
		{"duplicate effects", "<?php $values = ['abc', 'xyz']; $i = 0; substr($values[$i++], -1, 1);", 0, ""},
		{"magic source", "<?php class C { public string $s; } function f(C $c) { return substr($c->s, -1, 1); }", 0, ""},
		{"reference capture", "<?php function f(string $s) { $o = 0; $d = function () use (&$o) { $o -= 2; }; $d(); return substr($s, $o, 1); }", 0, ""},
		{"global offset", "<?php function f(string $s) { global $o; $o = 0; g(); return substr($s, $o, 1); }", 0, ""},
		{"static offset", "<?php function f(string $s) { static $o; $o = 0; g(); return substr($s, $o, 1); }", 0, ""},
		{"reference parameter", "<?php function f(string $s, int &$o) { $o = 0; g(); return substr($s, $o, 1); }", 0, ""},
		{"captured inside", "<?php $f = function (string $s) use (&$o) { $o = 0; g(); return substr($s, $o, 1); };", 0, ""},
		{"reference item", "<?php function f(string $s) { $o = 0; $a = ['k' => &$o]; g($a); return substr($s, $o, 1); }", 0, ""},
		{"reference assignment", "<?php function f(string $s) { $o = 0; $r = &$o; $r = -3; return substr($s, $o, 1); }", 0, ""},
		{"decremented twice", "<?php function f(string $s) { $o = 0; $o--; $o--; return substr($s, $o, 1); }", 0, ""},
		{"decremented in loop", "<?php function f(string $s) { $o = 0; while (g()) { echo substr($s, $o, 1); $o--; } }", 0, ""},
		{"top-level global", "<?php $s = 'abc'; $o = 0; g(); echo substr($s, $o, 1);", 0, ""},
		{"local offset", "<?php function f(string $s) { $o = 0; $c = function () use ($o) { return $o; }; $c(); return substr($s, $o, 1); }", 1, "<?php function f(string $s) { $o = 0; $c = function () use ($o) { return $o; }; $c(); return ($s[$o] ?? ''); }"},
		{"known zero", "<?php $s = 'abc'; substr($s, 0, 1);", 1, "<?php $s = 'abc'; ($s[0] ?? '');"},
		{"known negative one", "<?php function f(string $s) { $offset = -1; return substr($s, $offset, 1); }", 1, "<?php function f(string $s) { $offset = -1; return ($s[strlen($s) - 1] ?? ''); }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("audit.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "audit.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := tc.source
			if tc.fixed != "" {
				want = tc.fixed
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("audit.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}

func BenchmarkAuditSafety(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.ParseBest("audit.php", []byte("<?php $s = 'abc'; substr($s, 0, 1);"), syntax.Options{Version: phpversion.PHP85})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(f)
	}
}
