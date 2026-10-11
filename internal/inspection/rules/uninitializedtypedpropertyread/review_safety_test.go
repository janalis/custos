package uninitializedtypedpropertyread_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/uninitializedtypedpropertyread"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	const conf = "<?php class Conf{public array $opts;public ?Conf $next;public int $n;function items(){return [];}}\n"
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"isset element", conf + "isset((new Conf)->opts['k']);", 0},
		{"isset sub-property", conf + "isset((new Conf)->next->n);", 0},
		{"empty element", conf + "empty((new Conf)->opts['k']);", 0},
		{"coalesce element", conf + "echo (new Conf)->opts['k'] ?? 1;", 0},
		{"coalesce sub-property", conf + "echo (new Conf)->next->n ?? 1;", 0},
		{"coalesce nullsafe", conf + "echo (new Conf)->next?->n ?? 1;", 0},
		{"coalesce deep chain", conf + "echo (new Conf)->next->next->opts['a']['b'] ?? 1;", 0},
		{"coalesce parenthesized", conf + "echo ((new Conf)->opts)['k'] ?? 1;", 0},
		{"isset through call", conf + "isset((new Conf)->next->items()['k']);", 1},
		{"coalesce through call", conf + "echo (new Conf)->next->items()['k'] ?? 1;", 1},
		{"coalesce index position", conf + "echo [1][(new Conf)->n] ?? 1;", 1},
		{"coalesce dynamic name position", conf + "echo $o->{(new Conf)->n} ?? 1;", 1},
		{"coalesce right operand", conf + "echo null ?? (new Conf)->opts['k'];", 1},
		{"unguarded element", conf + "echo (new Conf)->opts['k'];", 1},
		{"other operator", conf + "echo (new Conf)->opts['k'] + 1;", 1},
		{"new static open class", "<?php class P{public int $v;function peek(){return (new static)->v;}}", 0},
		{"new static final class", "<?php final class P{public int $v;function peek(){return (new static)->v;}}", 1},
		{"new self open class", "<?php class P{public int $v;function peek(){return (new self)->v;}}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
