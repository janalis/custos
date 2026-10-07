package util

import (
	"sort"
	"strings"
	"testing"

	"custos/internal/syntax"
)

// completeValuesOf returns the texts of the complete value set of the first
// argument of the last call to probe() in src, or "?" when incomplete.
func completeValuesOf(t *testing.T, src string) string {
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
	vals, complete := PossibleValuesComplete(f, arg)
	if !complete {
		return "?"
	}
	var out []string
	for _, v := range vals {
		out = append(out, text(f, v))
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

func TestPossibleValuesComplete(t *testing.T) {
	for src, want := range map[string]string{
		`<?php probe('x');`:                                                                                                 "'x'",
		`<?php probe($a ? 'x' : ('y' ?? 'z'));`:                                                                             "'x' 'y' 'z'",
		`<?php function f(bool $a) { probe($a ? 'x' : 'y'); }`:                                                              "'x' 'y'",
		`<?php function f($m) { probe($m ?: 'x'); }`:                                                                        "?",
		`<?php function f() { $m = g(); probe($m ?: 'x'); }`:                                                                "'x' g()",
		`<?php function f($p = 'd') { probe($p); }`:                                                                         "?",
		`<?php function f($p = 'd') { $p = 'e'; probe($p); }`:                                                               "'e'",
		`<?php function f() { $p = 'a'; $p = 'b'; probe($p); }`:                                                             "'b'",
		`<?php function f($c) { $p = 'a'; if ($c) { $p = 'b'; } probe($p); }`:                                               "'a' 'b'",
		`<?php function f($c) { if ($c) { $p = 'b'; } probe($p); }`:                                                         "?",
		`<?php function f() { foreach ([1] as $p) {} probe($p); }`:                                                          "?",
		`<?php function f() { $p = 'a'; $p .= 'b'; probe($p); }`:                                                            "?",
		`<?php function f() { $p = 'a'; $g = function () use (&$p) {}; probe($p); }`:                                        "?",
		`<?php $top = 'x'; probe($top);`:                                                                                    "?",
		`<?php class C { const K = 'k'; function m() { probe(self::K); } }`:                                                 "'k'",
		`<?php class C { const K = 'k'; function m() { probe(static::K); } }`:                                               "?",
		`<?php final class C { const K = 'k'; function m() { probe(static::K); } }`:                                         "'k'",
		`<?php class C { function m() { probe(D::K); } }`:                                                                   "?",
		`<?php class C { private $p = 'd'; function m() { probe($this->p); } function n() { $this->p = 'e'; } }`:            "'d' 'e'",
		`<?php class C { public $p = 'd'; function m() { probe($this->p); } }`:                                              "'d'",
		`<?php class C { private $p; function m() { probe($this->p); } }`:                                                   "?",
		`<?php class C { private string $p; function __construct() { $this->p = 'c'; } function m() { probe($this->p); } }`: "'c'",
		`<?php class C { private $p = 'd'; function m() { $this->p .= 'x'; probe($this->p); } }`:                            "?",
		`<?php class C { private static $p = 's'; function m() { probe(self::$p); } }`:                                      "'s'",
		`<?php const G = 'g'; probe(G);`:                                                                                    "'g'",
		`<?php probe(UNKNOWN);`:                                                                                             "?",
		`<?php probe(null);`:                                                                                                "null",
	} {
		if got := completeValuesOf(t, src); got != want {
			t.Errorf("%s\n got %q, want %q", src, got, want)
		}
	}
}
