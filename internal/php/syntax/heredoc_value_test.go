package syntax

import (
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestEndlessStaticHeredoc(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		truth bool
	}{
		{"<<<E\nx\nE", true},
		{"<<<'E'\nx\nE", true},
		{"<<<\"E\"\nx\nE", true},
		{"<<<E\nE", false},
		{"<<<'E'\nE", false},
		{"<<<E\n0\nE", false},
		{"<<<'E'\n0\nE", false},
		{"<<<E\n\\x30\nE", false},
		{"<<<E\n\\060\nE", false},
		{"<<<'E'\n\\x30\nE", true},
		{"<<<E\n\\n\nE", true},
		{"b<<<E\n0\nE", false},
		{"B<<<'E'\nx\nE", true},
		{"<<<E\n  0\n  E", false},
		{"<<<'E'\n\t0\n\tE", false},
		{"<<<E\n  x\n  E", true},
		{"<<<E\n  x\n  y\n  E", true},
		{"<<<E\n \n  0\n  E", true},
		{"<<<E\r\n  0\r\n  E", false},
		{"<<<E\r  0\r  E", false},
		{"<<<E\r\n  x\r\n  y\r\n  E", true},
		{"<<<E\n\nE", false},
		{"<<<E\n\n\nE", true},
	} {
		f := parse(t, "<?php while ("+tc.raw+") {}", phpversion.PHP85)
		if len(f.Errors) != 0 {
			t.Fatalf("%q: %v", tc.raw, f.Errors)
		}
		loop := f.Stmts[0].(*While)
		if _, ok := loop.Cond.(*Literal); !ok {
			t.Fatalf("not static literal: %q", tc.raw)
		}
		if got := Endless(loop); got != tc.truth {
			t.Errorf("%q: %v want %v", tc.raw, got, tc.truth)
		}
		if _, ok := StringLiteralValue(tc.raw); ok {
			t.Fatalf("quoted API accepted heredoc %q", tc.raw)
		}
	}
	f := parse(t, "<?php while (<<<E\n$variable\nE) {}", phpversion.PHP85)
	if Endless(f.Stmts[0]) {
		t.Fatal("interpolation is not a constant")
	}
}

func TestStaticHeredocValueRejectsRecovery(t *testing.T) {
	for _, raw := range []string{"", "b", "text", "<<<", "<<<'E\n0\nE", "<<<E\n0", "<<<E\n0\nE;", "<<<E\n$variable\nE", "<<<E\nx\n  E", "<<<E\n x\n \tE"} {
		if value, ok := staticHeredocValueForVersion(raw, phpversion.Max); ok {
			t.Errorf("accepted %q as %q", raw, value)
		}
	}
}

func BenchmarkStaticHeredocTruth(b *testing.B) {
	for _, body := range []string{"0", "\\x30", "text", strings.Repeat("  text\n", 64)} {
		b.Run(body[:min(len(body), 12)], func(b *testing.B) {
			raw := "<<<E\n" + body + "\nE"
			b.ReportAllocs()
			for b.Loop() {
				staticHeredocValueForVersion(raw, phpversion.Max)
			}
		})
	}
}

func FuzzStaticHeredocValue(f *testing.F) {
	for _, raw := range []string{"<<<E\n0\nE", "<<<'E'\n\\0\nE", "<<<E\r\n  x\r\n  E", "<<<E\n$variable\nE", "<<<E\n\\"} {
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 65536 {
			return
		}
		staticHeredocValueForVersion(raw, phpversion.Max)
	})
}

func BenchmarkCachedHeredocEndless(b *testing.B) {
	f := Parse("bench.php", []byte("<?php while (<<<E\n"+strings.Repeat("text\n", 1024)+"E) {}"), Options{Version: phpversion.PHP85})
	loop := f.Stmts[0]
	Endless(loop)
	b.ReportAllocs()
	for b.Loop() {
		Endless(loop)
	}
}
