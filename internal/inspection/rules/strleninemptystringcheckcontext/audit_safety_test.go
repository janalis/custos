package strleninemptystringcheckcontext

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
		{"encoding effect", "<?php mb_strlen('a', encoding()) > 0;", 0, ""},
		{"shift encoding", "<?php mb_strlen(\"\\x1b(B\", 'ISO-2022-JP') === 0;", 0, ""},
		{"mutable encoding", "<?php if (mb_strlen($s)) {}", 0, ""},
		{"extra args", "<?php strlen('a', extra()) > 0;", 0, ""},
		{"byte length", "<?php strlen('a') > 0;", 1, "<?php 'a' !== '';"},
		{"negated target", "<?php $r = !!strlen($s);", 1, "<?php $r = !((string)$s === '');"},
		{"cast target", "<?php $r = (int) !strlen($s);", 1, "<?php $r = (int) ((string)$s === '');"},
		{"concat target", "<?php $r = !strlen($s) . 'x';", 1, "<?php $r = ((string)$s === '') . 'x';"},
		{"chained equality", "<?php $r = !strlen($a) === !strlen($b);", 2, "<?php $r = ((string)$a === '') === ((string)$b === '');"},
		{"comparison under equality", "<?php $r = strlen($a) > 0 === $f;", 1, "<?php $r = ((string)$a !== '') === $f;"},
		{"logical parent", "<?php $r = $f || !strlen($a);", 1, "<?php $r = $f || (string)$a === '';"},
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
	f := syntax.ParseBest("audit.php", []byte("<?php strlen('a') > 0;"), syntax.Options{Version: phpversion.PHP85})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(f)
	}
}
