package util

import (
	"sort"
	"strings"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

// valuesOf returns the texts of the possible values of the first argument
// of the last call to probe() in src.
func valuesOf(t *testing.T, src string) string {
	t.Helper()
	f := parse(t, src)
	var arg syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && CallLastName(c) == "probe" {
			args, _ := CallArgValues(c)
			arg = args[0]
		}
		return true
	})
	var out []string
	vals, known := PossibleValuesKnown(f, arg)
	if !known {
		return "?"
	}
	for _, v := range vals {
		out = append(out, text(f, v))
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

func TestPossibleValues(t *testing.T) {
	for src, want := range map[string]string{
		`<?php probe($a ? 'x' : ('y' ?? 'z'));`:                                                                            "'x' 'y' 'z'",
		`<?php function f($p = 1) { $p = $q = 2; probe($p); }`:                                                             "1 2",
		`<?php $top = 1; probe($top);`:                                                                                     "",
		`<?php class C { const K = 'k'; function m() { probe(self::K); } }`:                                                "'k'",
		`<?php class C { private $p = 'd'; function __construct() { $this->p = 'c'; } function m() { probe($this->p); } }`: "'c' 'd'",
		`<?php const G = 'g'; probe(G);`:                                                                                   "'g'",
		`<?php define('H', 'h'); probe(H);`:                                                                                "'h'",
		`<?php probe(UNKNOWN);`:                                                                                            "",
		`<?php probe(null);`:                                                                                               "null",
		`<?php probe(f() . 'x');`:                                                                                          "f() . 'x'",
		`<?php function f() { $v = 'a'; $g = function () { $v = 'b'; }; probe($v); }`:                                      "'a' 'b'",
		`<?php function f() { $n = 0; foreach ([1] as $x) { ++$n; } probe($n); }`:                                          "?",
		`<?php function f() { $n = 0; $g = function () use (&$n) { $n--; }; probe($n); }`:                                  "?",
		`<?php function f() { $s = 'P1D'; if (g()) { $s .= 'T2H'; } probe($s); }`:                                          "?",
		`<?php function f($o = null) { $o ??= 'x'; probe($o); }`:                                                           "?",
		`<?php function f(bool $c) { $s = 'a'; $s |= 4; probe($c ? 'b' : $s); }`:                                           "?",
		`<?php function f() { $t = 'a'; $u = $t; $t += 1; probe($u); }`:                                                    "?",
		`<?php function f() { $n = 0; $m++; $n = $m . 'x'; probe($n); }`:                                                   "$m . 'x' 0",
		`<?php function f() { $n = 0; $other++; $arr[$n] .= 'x'; probe($n); }`:                                             "0",
		`<?php $n = 0; $n++; probe($n);`:                                                                                   "",
	} {
		if got := valuesOf(t, src); got != want {
			t.Errorf("%s:\n got %q\nwant %q", src, got, want)
		}
	}
}

// BenchmarkPossibleValuesKnown measures discovery on a variable of a large
// function body (assignment scan + unstable-variable scan).
func BenchmarkPossibleValuesKnown(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php function f($p = 1) {\n")
	for i := 0; i < 500; i++ {
		sb.WriteString("$a = $b . 'x'; $c[$a] = foo($a, $b ?? 2); $d .= $c[$a];\n")
	}
	sb.WriteString("$p = 3; probe($p); }\n")
	f := syntax.Parse("b.php", []byte(sb.String()), syntax.Options{Version: phpver.Max})
	var arg syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && CallLastName(c) == "probe" {
			args, _ := CallArgValues(c)
			arg = args[0]
		}
		return true
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, known := PossibleValuesKnown(f, arg); !known {
			b.Fatal("unexpected unknown result")
		}
	}
}

func TestPossibleValuesEdges(t *testing.T) {
	many := strings.Repeat
	for src, want := range map[string]string{
		`<?php function f() { probe($$x); }`:                                                                     "",
		`<?php class C { function m() { probe($this->{$x}); } }`:                                                 "",
		`<?php class C { const K = 'k'; function m() { probe(C::{$x}); } }`:                                      "",
		`<?php function f($o) { probe($o::K); }`:                                                                 "",
		`<?php class C { const K = 'k'; } function f() { probe(C::K); }`:                                         "'k'",
		`<?php class C { const K = 'k'; function m() { probe(static::K); } }`:                                    "'k'",
		`<?php function f() { probe(Missing::K); }`:                                                              "",
		`<?php define('X'); probe(X);`:                                                                           "",
		`<?php function f() { ` + many(`$v = 1; `, maxPossibleValues+1) + `probe($v); }`:                         "?",
		`<?php class C { function m() { ` + many(`$this->p = 1; `, maxPossibleValues+1) + `probe($this->p); } }`: "?",
	} {
		if got := valuesOf(t, src); got != want {
			t.Errorf("%.80s:\n got %q\nwant %q", src, got, want)
		}
	}
	f := parse(t, `<?php 1;`)
	if vals, known := PossibleValuesKnown(f, nil); vals != nil || !known {
		t.Fatalf("nil expression: %v %v", vals, known)
	}
	if p, b := ScopeParts(nil); p != nil || b != nil {
		t.Fatal("ScopeParts of a non-function")
	}
}

func TestPossibleValuesAbstractScope(t *testing.T) {
	// A variable in an abstract method's parameter default: no body to scan.
	if got := valuesOf(t, `<?php interface I { function f($a = probe($b)); }`); got != "" {
		t.Fatalf("got %q", got)
	}
}
