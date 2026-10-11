package intltwelvehourfieldusedfortwentyfourhourtime

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
	const head = "<?php function f(IntlCalendar $c) { $h = 12; "
	for _, tc := range []struct {
		name, source string
		fixed        string
	}{
		{"decrement", head + "$h--; $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"pre-increment", head + "++$h; $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"nested increment", head + "$x = $h++ + 1; $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"reference capture", head + "$g = function () use (&$h) { $h = 3; }; $g(); $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"catch binding", head + "try { g(); } catch (Exception $h) {} $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"compound", head + "$h -= 3; $c->set(IntlCalendar::FIELD_HOUR, $h); }", ""},
		{"top-level decrement", "<?php $c = IntlCalendar::createInstance(); $h = 12; $h--; $c->set(IntlCalendar::FIELD_HOUR, $h);", ""},
		{"arrow scope decrement", "<?php fn(IntlCalendar $c) => [$h = 12, $h--, $c->set(IntlCalendar::FIELD_HOUR, $h)];", ""},
		{
			"unchanged local", head + "$c->set(IntlCalendar::FIELD_HOUR, $h); }",
			head + "$c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }",
		},
		{
			"value capture", head + "$g = function () use ($h) { $h--; }; try { $g(); } catch (Exception) {} $c->set(IntlCalendar::FIELD_HOUR, $h); }",
			head + "$g = function () use ($h) { $h--; }; try { $g(); } catch (Exception) {} $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }",
		},
		{
			"other variable", head + "$i = 0; $i++; $r = function () use (&$i) {}; try { g(); } catch (Exception $e) {} $c->set(IntlCalendar::FIELD_HOUR, $h); }",
			head + "$i = 0; $i++; $r = function () use (&$i) {}; try { g(); } catch (Exception $e) {} $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }",
		},
		{
			"arrow copy", head + "$g = fn() => $h++; $c->set(IntlCalendar::FIELD_HOUR, $h); }",
			head + "$g = fn() => $h++; $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h); }",
		},
		{
			"other function", "<?php function g() { $h = 0; $h++; } $c = IntlCalendar::createInstance(); $h = 12; $c->set(IntlCalendar::FIELD_HOUR, $h);",
			"<?php function g() { $h = 0; $h++; } $c = IntlCalendar::createInstance(); $h = 12; $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, $h);",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			want, findings := tc.source, 0
			if tc.fixed != "" {
				want, findings = tc.fixed, 1
			}
			if got := e.Analyze(f); len(got) != findings {
				t.Fatalf("findings: got %v, want %d", got, findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
			if len(syntax.ParseBest("review.php", res.Source, opt).Errors) != 0 {
				t.Fatal("fix introduced syntax errors")
			}
		})
	}
}
