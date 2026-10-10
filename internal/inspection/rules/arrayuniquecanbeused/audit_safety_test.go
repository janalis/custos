package arrayuniquecanbeused

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
		{"numeric string keys", "<?php array_keys(array_count_values(['1']));", 0, ""},
		{"unsupported values", "<?php count(array_count_values(['01', '1', 1, 1.5, false]));", 0, ""},
		{"unknown values", "<?php function f(array $a) { return count(array_count_values($a)); }", 0, ""},
		{"safe keys", "<?php array_keys(array_count_values(['alpha', 'beta', 'alpha']));", 1, "<?php array_values(array_unique(['alpha', 'beta', 'alpha']));"},
		{"safe ints", "<?php array_keys(array_count_values([1, 2, 1]));", 1, "<?php array_values(array_unique([1, 2, 1]));"},
		{"safe count", "<?php count(array_count_values(['1', 1]));", 1, "<?php count(array_unique(['1', 1]));"},
		{"empty", "<?php array_keys(array_count_values([]));", 1, "<?php array_values(array_unique([]));"},
		{"known locals", "<?php function f() { $a = ['a', 'b']; return array_keys(array_count_values($a)); }", 1, "<?php function f() { $a = ['a', 'b']; return array_values(array_unique($a)); }"},
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
	f := syntax.ParseBest("audit.php", []byte("<?php array_keys(array_count_values(['alpha', 'beta', 'alpha']));"), syntax.Options{Version: phpversion.PHP85})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(f)
	}
}
