package poweroperatorcanbeused

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
		{"yield base", "<?php function g() { $r = pow(yield 3, 2); }", 1, "<?php function g() { $r = (yield 3) ** 2; }"},
		{"yield exponent", "<?php function g() { return pow(2, yield 3) ? 1 : 0; }", 1, "<?php function g() { return 2 ** (yield 3) ? 1 : 0; }"},
		{"print base", "<?php $r = pow(print 3, 2);", 1, "<?php $r = (print 3) ** 2;"},
		{"include base", "<?php $r = pow(include __DIR__ . '/inc.php', 2);", 1, "<?php $r = (include __DIR__ . '/inc.php') ** 2;"},
		{"require exponent", "<?php $r = pow(2, require $f);", 1, "<?php $r = 2 ** (require $f);"},
		{"throw base", "<?php $r = pow(throw $e, 2);", 1, "<?php $r = (throw $e) ** 2;"},
		{"arrow base", "<?php $r = pow(fn() => 2, 2);", 1, "<?php $r = (fn() => 2) ** 2;"},
		{"array access", "<?php $r = pow(2, 3)[0] ?? 'none';", 1, "<?php $r = (2 ** 3)[0] ?? 'none';"},
		{"property", "<?php $r = pow($a, 2)->p ?? 'np';", 1, "<?php $r = ($a ** 2)->p ?? 'np';"},
		{"nullsafe method", "<?php $r = pow($a, 2)?->m();", 1, "<?php $r = ($a ** 2)?->m();"},
		{"call result", "<?php $r = pow($a, 2)();", 1, "<?php $r = ($a ** 2)();"},
		{"static call", "<?php $r = pow($a, 2)::make();", 1, "<?php $r = ($a ** 2)::make();"},
		{"static property", "<?php $r = pow($a, 2)::$p;", 1, "<?php $r = ($a ** 2)::$p;"},
		{"class constant", "<?php $r = pow($a, 2)::C;", 1, "<?php $r = ($a ** 2)::C;"},
		{"instanceof", "<?php $r = pow($a, !$b) instanceof GMP;", 1, "<?php $r = ($a ** !$b) instanceof GMP;"},
		{"clone", "<?php $r = clone pow($g, 2);", 1, "<?php $r = clone ($g ** 2);"},
		{"index operand", "<?php $r = $t[pow($a, 2)];", 1, "<?php $r = $t[$a ** 2];"},
		{"method argument", "<?php $r = $o->m(pow($a, 2));", 1, "<?php $r = $o->m($a ** 2);"},
		{"cast parent", "<?php $r = (int)pow($a, 2);", 1, "<?php $r = (int)$a ** 2;"},
		{"comment before base", "<?php $r = pow(/* base */ $x, 3);", 1, ""},
		{"comment before exponent", "<?php $r = pow($x, /* exp */ 3);", 1, ""},
		{"line comment after base", "<?php $r = pow($x // base\n, 3);", 1, ""},
		{"line comment after exponent", "<?php $r = pow($x, 3 # exp\n);", 1, ""},
		{"comment before list", "<?php $r = pow /* call */ ($x, 3);", 1, ""},
		{"comment inside argument", "<?php $r = pow($x /* a */ + 1, 3);", 1, "<?php $r = ($x /* a */ + 1) ** 3;"},
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
