package intlcalendarsecondspassedasmilliseconds

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
	const pre = "<?php function f(IntlCalendar $c) { "
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"reference capture", pre + "$t = time(); $f = function () use (&$t) { $t *= 1000; }; $f(); $c->setTime($t); }", 0, ""},
		{"reference capture of source", pre + "$t = time(); $f = function () use (&$t) { $t *= 1000; }; $f(); $u = $t; $c->setTime($u); }", 0, ""},
		{"value capture", pre + "$t = time(); $f = function () use ($t) { $t *= 1000; }; $f(); $c->setTime($t); }", 1, ""},
		{"other reference capture", pre + "$t = time(); $f = function () use (&$o) { $o = 1; }; $f(); $c->setTime($t); }", 1, ""},
		{"copy before capture", pre + "$t = time(); $u = $t; $u = $u; $c->setTime($u); }", 1, ""},
		{"own reference capture", "<?php $f = function (IntlCalendar $c) use (&$t) { $t = time(); $c->setTime($t); };", 1, ""},
		{"nested scopes", pre + "$t = time(); function g() { $f = function () use (&$t) {}; } class K { function m() { $f = function () use (&$t) {}; } } $a = fn() => function () use (&$t) {}; $c->setTime($t); }", 1, ""},
		{"catch binding", pre + "$t = time(); try { g(); } catch (E $t) {} $c->setTime($t); }", 0, ""},
		{"catch without binding", pre + "$t = time(); try { g(); } catch (E) {} $c->setTime($t); }", 1, ""},
		{"file level", "<?php $c = IntlCalendar::createInstance(); $t = time(); $c->setTime($t);", 1, ""},
		{"file level other global", "<?php function z() { global $o, $$n; } $c = IntlCalendar::createInstance(); $t = time(); z(); $c->setTime($t);", 1, ""},
		{"global declaration", "<?php function z() { global $t; $t *= 1000; } $c = IntlCalendar::createInstance(); $t = time(); z(); $c->setTime($t);", 0, ""},
		{"globals array", "<?php function z() { $GLOBALS['t'] *= 1000; } $c = IntlCalendar::createInstance(); $t = time(); z(); $c->setTime($t);", 0, ""},
		{"file level capture", "<?php $c = IntlCalendar::createInstance(); $t = time(); $f = function () use (&$t) { $t *= 1000; }; $f(); $c->setTime($t);", 0, ""},
		{"parenthesized comment", pre + "$c->setTime((/* wall clock */ time())); }", 1, pre + "$c->setTime((/* wall clock */ time() * 1000)); }"},
		{"nested parentheses", pre + "$c->setTime(( (time()) /* s */ )); }", 1, pre + "$c->setTime(( (time() * 1000) /* s */ )); }"},
		{"comment in call", pre + "$c->setTime(timestamp: \\time( /* now */ ) /* s */); }", 1, pre + "$c->setTime(timestamp: (\\time( /* now */ ) * 1000) /* s */); }"},
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
