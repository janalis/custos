package nonsecureuniqidusage

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	const closure = "function ($value = '') { return uniqid((string) $value, true); }"
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
		fixed        string
	}{
		{"parameter default", "<?php function f($l, $cb = 'uniqid') { return array_map($cb, $l); }", 1, ""},
		{"fallback", "<?php function f($l, $o) { $cb = $o ?: 'uniqid'; return array_map($cb, $l); }", 1, ""},
		{"dynamic list", "<?php call_user_func_array('uniqid', $args);", 1, ""},
		{"empty list", "<?php call_user_func_array('uniqid', []);", 1, ""},
		{"keyed list", "<?php call_user_func_array('uniqid', ['prefix' => 'x']);", 1, ""},
		{"integer key", "<?php call_user_func_array('uniqid', [3 => 'x']);", 1, ""},
		{"named forward", "<?php call_user_func('uniqid', prefix: 'x');", 1, ""},
		{"unpacked forward", "<?php call_user_func('uniqid', ...$a);", 1, ""},
		{"forwarded entropy", "<?php call_user_func('uniqid', 'p_', true);", 0, ""},
		{"forwarded entropy list", "<?php call_user_func_array('uniqid', ['p_', true]);", 0, ""},
		{"forwarded entropy key", `<?php call_user_func_array('uniqid', ["more_entropy" => true]);`, 0, ""},
		{"two arrays", "<?php array_map('uniqid', $a, $b);", 0, ""},
		{"single list", "<?php call_user_func_array('uniqid', ['x']);", 1, "<?php call_user_func_array(" + closure + ", ['x']);"},
		{"single value", "<?php call_user_func('uniqid', 5);", 1, "<?php call_user_func(" + closure + ", 5);"},
		{"empty comment", "<?php uniqid(/* c */);", 1, "<?php uniqid('', true/* c */);"},
		{"line comment", "<?php uniqid('a' // c\n);", 1, "<?php uniqid('a', true // c\n);"},
		{"trailing comma", "<?php uniqid('a' /* c */ ,);", 1, "<?php uniqid('a', true /* c */ ,);"},
		{"named trailing comma", "<?php uniqid(prefix: 'q',);", 1, "<?php uniqid(prefix: 'q', more_entropy: true,);"},
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
