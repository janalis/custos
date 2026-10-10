package arraypushmissuse

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
		{"sparse", "<?php $a = [2 => 'old']; $a[count($a)] = 'new';", 0, ""},
		{"deleted next index", "<?php $a = ['old', 'last']; unset($a[1]); $a[count($a)] = 'new';", 0, ""},
		{"unknown", "<?php function f(array $a) { $a[count($a)] = 'new'; }", 0, ""},
		{"packed", "<?php $a = ['old']; $a[count($a)] = 'new';", 1, ""},
		{"empty", "<?php $a = []; $a[count($a)] = 'new';", 1, ""},
		{"not statement", "<?php $a = []; $v = ($a[count($a)] = 'new');", 0, ""},
		{"side effect value", "<?php $a = []; $a[count($a)] = mutate($a);", 0, ""},
		{"property", "<?php $a->x[count($a->x)] = 'new';", 0, ""},
		{"reference", "<?php $a = &$b; $a[count($a)] = 'new';", 0, ""},
		{"spread", "<?php $a = [...$b]; $a[count($a)] = 'new';", 0, ""},
		{"referenced item", "<?php $a = [&$b]; $a[count($a)] = 'new';", 0, ""},
		{"previous non assignment", "<?php ping(); $a[count($a)] = 'new';", 0, ""},
		{"previous no expr", "<?php echo 'ok'; $a[count($a)] = 'new';", 0, ""},
		{"different variable", "<?php $b = []; $a[count($a)] = 'new';", 0, ""},
		{"compound previous", "<?php $a += []; $a[count($a)] = 'new';", 0, ""},
		{"nonliteral", "<?php $a = source(); $a[count($a)] = 'new';", 0, ""},
		{"variable variable", "<?php $$a[count($$a)] = 'new';", 0, ""},
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
	f := syntax.ParseBest("audit.php", []byte("<?php $a = ['old']; $a[count($a)] = 'new';"), syntax.Options{Version: phpversion.PHP85})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(f)
	}
}
