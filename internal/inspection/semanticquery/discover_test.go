package semanticquery

import (
	"sort"
	"strings"
	"testing"

	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func discoverOf(t *testing.T, src string) string {
	t.Helper()
	f := Parse(t, src)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
	var arg syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(c) == "probe" {
			args, _ := astquery.CallArgValues(c)
			arg = args[0]
		}
		return true
	})
	var out []string
	vals, known := DiscoverValuesKnown(env, arg)
	if !known {
		return "?"
	}
	for _, v := range vals {
		out = append(out, Text(f, v))
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

func TestDiscoverValues(t *testing.T) {
	for src, want := range map[string]string{
		`<?php probe($a ? 'x' : ('y' ?? 'z'));`:                "'x' 'y' 'z'",
		`<?php probe($a ?: 'z');`:                              "'z'",
		`<?php function f($p = 1) { $p = $q = 2; probe($p); }`: "1 2",
		`<?php $top = 1; probe($top);`:                         "",
		`<?php class P { const K = 'k'; } class C extends P { function m() { probe(self::K); } }`:                          "'k'",
		`<?php class C { const K = 'k'; } function f() { probe(C::K); }`:                                                   "'k'",
		`<?php class C { private $p = 'd'; function __construct() { $this->p = 'c'; } function m() { probe($this->p); } }`: "'c' 'd'",
		`<?php class C { static $s = 's'; function m() { probe(static::$s); } }`:                                           "'s'",
		`<?php class C { public $p = 'd'; } function f(C $c) { $c->p = 'e'; probe($c->p); }`:                               "'d' 'e'",
		`<?php const G = 'g'; probe(G);`:      "'g'",
		`<?php define('H', A | B); probe(H);`: "A | B",
		`<?php probe(JSON_THROW_ON_ERROR);`:   "JSON_THROW_ON_ERROR",
		`<?php probe(UNKNOWN);`:               "",
		`<?php probe(null);`:                  "null",
		`<?php probe(X::class);`:              "X::class",
		`<?php function f() { $v = 'a'; $g = function () { $v = 'b'; }; probe($v); }`:     "'a' 'b'",
		`<?php function f() { $n = 0; foreach ([1] as $x) { ++$n; } probe($n); }`:         "?",
		`<?php function f() { $n = 0; $g = function () use (&$n) { $n--; }; probe($n); }`: "?",
		`<?php function f() { $s = 'P1D'; if (g()) { $s .= 'T2H'; } probe($s); }`:         "?",
		`<?php function f($o = null) { $o ??= 'x'; probe($o); }`:                          "?",
		`<?php function f(bool $c) { $s = 'a'; $s |= 4; probe($c ? 'b' : $s); }`:          "?",
		`<?php function f() { $t = 'a'; $u = $t; $t += 1; probe($u); }`:                   "?",
		`<?php function f() { $n = 0; $m++; $n = $m . 'x'; probe($n); }`:                  "$m . 'x' 0",
		`<?php function f() { $n = 0; $other++; $arr[$n] .= 'x'; probe($n); }`:            "0",
		`<?php $n = 0; $n++; probe($n);`:                                                  "",
	} {
		if got := discoverOf(t, src); got != want {
			t.Errorf("%s:\n got %q\nwant %q", src, got, want)
		}
	}
}

func TestFunctionDecl(t *testing.T) {
	f := Parse(t, `<?php namespace A; if (true) { function g() { return 1; } }`)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	d := FunctionDecl(f, ix.Function(`A\g`, 0))
	if d == nil || d.Name.Value != "g" {
		t.Fatalf("got %v", d)
	}
	if FunctionDecl(f, ix.Function("strlen", 0)) != nil {
		t.Fatal("stub function must not resolve to a node")
	}
}

func TestDiscoverValuesEdges(t *testing.T) {
	many := func(n int) string { return strings.Repeat(`$this->p = 1; `, n) }
	for src, want := range map[string]string{
		`<?php class C {} function f(C $c) { probe($c->q); }`:                                                       "",
		`<?php class C { const K = 1; } function f() { probe(C::{$x}); }`:                                           "",
		`<?php function f($o) { probe($o::K); }`:                                                                    "",
		`<?php class C {} function f() { probe(C::NOPE); }`:                                                         "",
		`<?php enum E { case A; } function f() { probe(E::A); }`:                                                    "",
		`<?php function f() { probe(\DateTime::ATOM); }`:                                                            "",
		`<?php class C { public $p; function m() { ` + many(flowquery.MaxPossibleValues+1) + `probe($this->p); } }`: "?",
		`<?php class C { public $p; function m() { ` + many(flowquery.MaxAssignScan+1) + `probe($this->p); } }`:     "?",
	} {
		if got := discoverOf(t, src); got != want {
			t.Errorf("%.80s:\n got %q\nwant %q", src, got, want)
		}
	}
	f := Parse(t, `<?php 1;`)
	env := infer.NewEnv(f, names.New(f), index.New(stubs.Index()), phpversion.PHP84)
	if ResolveConstant(env, nil) != nil {
		t.Fatal("ResolveConstant(nil)")
	}
}
