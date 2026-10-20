package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

// nativeReads types the argument of every `u(...)` call with the native
// environment.
func nativeReads(t *testing.T, src string) []string {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
				got = append(got, env.TypeOf(c.Args.Args[0].(*syntax.Arg).Value).ShapeString())
			}
		}
		return true
	})
	return got
}

func checkReads(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A case, or a constant aliasing one, is typed with that case; strict
// comparisons with a case keep or remove it, and an enum method narrows
// $this the same way.
func TestEnumCaseNarrowing(t *testing.T) {
	checkReads(t, nativeReads(t, `<?php
namespace N;
enum E {
    case A; case B; case C;
    const X = self::A;
    const Y = E::X;
    const Z = \N\E::B;
    const W = Other::A;
    const V = 3;
    const L = self::M;
    const M = self::L;
    const Q = self::Missing;
    function own() {
        if ($this === self::A) { return; }
        u($this);
    }
}
enum Other { case A; }
class K { const A = 1; }
function u($x) {}
function f(E $e, ?E $n, $m, int|E $i, $name) {
    u(E::A);
    u(E::Y);
    u(E::Z);
    u(E::W);
    u(E::V);
    u(E::L);
    u(E::Q);
    u(E::{$name});
    u(K::A);
    if ($e === E::A) { u($e); }
    if ($e !== E::A) { u($e); }
    if ($e !== E::A && $e !== E::B) { u($e); }
    if ($e === E::A || $e === E::C) { u($e); }
    if ($n === E::B) { u($n); }
    if ($m === E::B) { u($m); }
    if ($i === E::X) { u($i); } else { u($i); }
    if ($e === Other::A) { u($e); }
    $k = E::B;
    u($k);
    if ($e === E::{$name}) { u($e); }
}`), []string{
		`\N\E::B|\N\E::C`,
		`\N\E::A`, `\N\E::A`, `\N\E::B`, "?unknown", "int", "?unknown", "?unknown", "?unknown", "int",
		`\N\E::A`, `\N\E::B|\N\E::C`, `\N\E::C`, `\N\E::A|\N\E::C`, `\N\E::B`, "?unknown",
		`\N\E::A`, `\N\E::B|\N\E::C|int`, "?unknown", `\N\E::B`, `\N\E`,
	})
}

// Where declarations are proofs, a fact about a local ends where the local
// may change unseen: a write through a reference bound to it, a call that
// may take it by reference, any call for a by-reference parameter, and
// eval().
func TestNativeLocalFactsEnd(t *testing.T) {
	checkReads(t, nativeReads(t, `<?php
enum E { case A; case B; }
class K { function m($v) {} }
class Plain {}
function u($x) {}
function g() {}
function byVal($v) {}
function f(E $e, E &$r, $o, K $k, string $cls, $meth) {
    if ($e === E::A) { return; }
    u($e);
    $p = &$e; u($e);
    if ($e === E::A) { return; } u($e);
    $p = E::A; u($e);
    $p = E::B; u($p);
    if ($r === E::A) { return; } u($r);
    g(); u($r);
}
function h(E $e, $o, K $k, string $cls, $meth) {
    if ($e === E::A) { return; } byVal($e); u($e);
    $k->m($e); u($e);
    missing($e); u($e);
}
function i(E $e, $o, K $k, string $cls, $meth) { if ($e === E::A) { return; } $o->m($e); u($e); }
function j(E $e, K $k) { if ($e === E::A) { return; } $k->missing($e); u($e); }
function k(E $e, K $k, $meth) { if ($e === E::A) { return; } $k->$meth($e); u($e); }
function l(E $e, $s) { if ($e === E::A) { return; } K::$s($e); u($e); }
function m(E $e, $cls) { if ($e === E::A) { return; } new $cls($e); u($e); }
function n(E $e) { if ($e === E::A) { return; } new Plain($e); u($e); }
function o(E $e) { if ($e === E::A) { return; } new Unknown($e); u($e); }
function p(E $e) { if ($e === E::A) { return; } eval('$e = E::A;'); u($e); }
function q(E $e) { $a = [&$e]; u($e); }
function r(array $a) { foreach ($a as &$v) { u($v); } }
function s(E $e) { $f = function () use (&$e) {}; if ($e === E::A) { return; } u($e); $f(); u($e); }
function t(E $e) { $p = &$e; $e = E::A; u($p); }
function v($o) { $o->m($x); u($x); }
function w(E $e) { if ($e === E::A) { return; } $f = function () use (&$e) {}; $f(); u($e); }
`), []string{
		`\E::B`, `\E`, `\E::B`, `\E`, `\E::B`, `\E::B`, `\E`,
		`\E::B`, `\E::B`, `\E`,
		`\E`, `\E`, `\E`, `\E`, `\E`, `\E::B`, `\E`, "?unknown",
		"?unknown", "?unknown", "?unknown", "?unknown", `\E`, "?unknown", "?unknown",
	})
}

// The byte length of a known string is a known integer for comparisons, and
// a local that a call does not touch keeps its only assignment even when
// passed to a method resolved late; NarrowedVarAt asks it where the local
// is not read.
func TestKnownLengthAndVarAt(t *testing.T) {
	src := `<?php
$f = new NumberFormatter("en_US", 1);
$s = "37tail";
$n = $f->parse($s, 1, $end);
if ($end === strlen($s)) { u($n); }
if ($end === mb_strlen("ab", '8bit')) { u($n); }
if ($end === mb_strlen("ab", 'UTF-8')) { u($n); }
if ($end === mb_strlen("ab", $enc)) { u($n); }
if ($end === strlen($s, 1)) { u($n); }
if ($end === strlen(...$a)) { u($n); }
if ($end === strlen($g)) { u($n); }
if ($end === strlen(1)) { u($n); }
if ($end === $f->strlen($s)) { u($n); }
if ($end === mb_strlen("ab")) { u($n); }
if ($end === $name("ab")) { u($n); }
if ($end === strlen()) { u($n); }
if ($end === strtolower("ab")) { u($n); }
if ($end === strlen($$k)) { u($n); }
$g = "abc"; $f->format($g); if ($end === strlen($g)) { u($n); }
u($n);
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	seed := types.Ints(types.IntRange{Lo: 0, Hi: 9})
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
				got = append(got, env.NarrowedVarAt("end", c, seed).ShapeString())
			}
		}
		return true
	})
	checkReads(t, got, []string{
		"int<6, 6>", "int<2, 2>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>",
		"int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<0, 9>", "int<3, 3>", "int<0, 9>",
	})
	if !env.NarrowedVarAt("", f.Stmts[0].(*syntax.ExprStmt).Expr, seed).IsUnknown() || !env.NarrowedVarAt("this", f.Stmts[0].(*syntax.ExprStmt).Expr, seed).IsUnknown() {
		t.Error("no name, or $this, is no local")
	}
}
