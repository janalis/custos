package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// Stub PHPDoc never widens a builtin's version-resolved native return.
func TestBuiltinDocDoesNotWiden(t *testing.T) {
	src := `<?php
function f(string $s) {
    t('substr', substr($s, 1));
    t('split', str_split($s));
    t('explode', explode(',', $s));
    t('cb', array_map('substr', [$s], [1]));
}
`
	checkVer(t, phpversion.PHP81, false, src, map[string]string{"substr": "string", "split": "string[]", "explode": "string[]", "cb": "non-empty string[]"})
	checkVer(t, phpversion.PHP74, false, src, map[string]string{"substr": "false|string", "split": "false|string[]"})
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP81})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP81))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
				want := map[string]string{"'substr'": "string", "'split'": "array|string[]", "'explode'": "array", "'cb'": "array"}[lit.Raw]
				if got != want {
					t.Errorf("T-rules %s: got %s want %s", lit.Raw, got, want)
				}
			}
		}
		return true
	})
}

// A user declaration's doc may widen its declared type (memberType picks
// the declared one); a builtin's never does.
func TestUserDocVsDeclared(t *testing.T) {
	checkAnywhere(t, `<?php
/** @return string|false */
function u(): string { return ''; }
function f() { t('user', u()); t('cb', array_map('u', [1])); t('strlen', strlen('x')); t('ob', ob_get_clean()); }
`, map[string]string{"user": "string", "cb": "string[]{0: string}", "strlen": "int", "ob": "false|string"})
}

// A method without return type keeps the return type of the method it
// overrides; declared `: mixed` is never replaced by the body.
func TestInheritedReturnSignature(t *testing.T) {
	checkAnywhere(t, `<?php
interface Source { public function value(): mixed; }
abstract class Base { /** @return int */ abstract public function count(); }
final class Impl extends Base implements Source {
    public function value() { return 1.5; }
    public function count() { return 'x'; }
    public function own() { return 2; }
    public function m(): mixed { return [1]; }
}
final class J implements \JsonSerializable { public function jsonSerialize() { return [1]; } }
function f(Impl $i, J $j) {
    t('value', $i->value());
    t('count', $i->count());
    t('own', $i->own());
    t('mixed', $i->m());
    t('json', $j->jsonSerialize());
}
`, map[string]string{"value": "mixed", "count": "int", "own": "int", "mixed": "mixed", "json": "mixed"})
}

// A documented `void` return contradicted by the body is dropped.
func TestVoidDocContradicted(t *testing.T) {
	checkAnywhere(t, `<?php
/** @return void */
function pair($t) { return [1, 2]; }
/** @return void */
function nothing() { $f = function () { return 1; }; return; }
final class K {
    /** @return void */
    public function m() { return 'x'; }
    /** @return void */
    public function v() { }
}
function f(K $k) { t('func', pair(1)); t('closure', nothing()); t('method', $k->m()); t('really', $k->v()); }
`, map[string]string{"func": "int[]{0: int, 1: int}", "closure": "void", "method": "string", "really": "void"})
}
