package flowquery

import (
	"sort"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestPossibleValuesReaching(t *testing.T) {
	for src, want := range map[string]string{
		`<?php function f($p = 'a') { probe($p); }`:                         "'a'",
		`<?php function f($p = 'a') { $p = 'b'; probe($p); }`:               "'b'",
		`<?php function f() { $x = 'a'; $x = 'b'; probe($x); }`:             "'b'",
		`<?php function f($c) { $x = $c ? 'a' : 'b'; probe($x ?? 'z'); }`:   "'a' 'b' 'z'",
		`<?php function f() { $x = 'a'; $x .= 'b'; probe($x); }`:            "?",
		`<?php class C { const K = 'k'; function f() { probe(self::K); } }`: "'k'",
	} {
		f := Parse(t, src)
		vals, known := PossibleValuesReaching(f, ProbeArg(f))
		got := "?"
		if known {
			var out []string
			for _, v := range vals {
				out = append(out, Text(f, v))
			}
			sort.Strings(out)
			got = strings.Join(out, " ")
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", src, got, want)
		}
	}
}

func TestPossibleValuesReachingEdges(t *testing.T) {
	for src, want := range map[string]bool{
		`<?php function f() { probe($$x); }`:                                               true,
		`<?php class C { function m() { $n = 1; $n++; $this->p = $n; probe($this->p); } }`: false,
	} {
		f := Parse(t, src)
		if vals, known := PossibleValuesReaching(f, ProbeArg(f)); known != want || len(vals) != 0 {
			t.Errorf("%s: got %v %v", src, vals, known)
		}
	}
}

func TestShortTernaryValueDiscovery(t *testing.T) {
	f := Parse(t, `<?php function f() { $x = 'extract' ?: 'compact'; probe($x ?: 'parse_str'); }`)
	for _, discover := range []func(*syntax.File, syntax.Expr) ([]syntax.Expr, bool){PossibleValuesKnown, PossibleValuesReaching, PossibleValuesComplete} {
		values, known := discover(f, ProbeArg(f))
		var got []string
		for _, value := range values {
			got = append(got, Text(f, value))
		}
		sort.Strings(got)
		if !known || strings.Join(got, " ") != "'compact' 'extract' 'parse_str'" {
			t.Fatalf("%v, %v", got, known)
		}
	}
}

func BenchmarkShortTernaryDiscovery(b *testing.B) {
	f := Parse(b, `<?php function f() { $x = 'extract' ?: 'compact'; probe($x ?: 'parse_str'); }`)
	arg := ProbeArg(f)
	b.ReportAllocs()
	for b.Loop() {
		PossibleValuesReaching(f, arg)
	}
}
