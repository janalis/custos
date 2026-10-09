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

// Every assignment replaces the variable's value: `.=` yields a string
// whatever the variable held.
func TestCompoundAssignmentReplaces(t *testing.T) {
	checkAnywhere(t, `<?php
function f($f, int $i) {
    $h = file_get_contents($f);
    $h .= 'x';
    t('concat', $h);
    $n = null;
    $n .= 'x';
    t('concatNull', $n);
    $m = null;
    $m = $m . 'x';
    t('concatExpr', $m);
    $k = '1';
    $k += $i;
    t('plus', $k);
    $r = 'a';
    $s = 'b';
    $r = &$s;
    t('byRef', $r);
}
`, map[string]string{"concat": "string", "concatNull": "string", "concatExpr": "string", "plus": "?unknown", "byRef": "string"})
}

// `$this`, `self`, `static` inside a trait are the unknown using class.
func TestTraitSelfUnknown(t *testing.T) {
	checkAnywhere(t, `<?php
trait T {
    public function f() {
        t('this', $this);
        t('new', new static());
        t('newSelf', new self());
        t('clone', clone $this);
        t('call', self::g());
    }
    public static function g(): static {}
}
class C { use T; public function h() { t('class', $this); } }
`, map[string]string{"this": "?unknown", "new": "?unknown", "newSelf": "?unknown", "clone": "?unknown", "call": "?unknown", "class": `\C`})
}

// A string failing is_numeric() is still a string.
func TestNegatedIsNumericOnString(t *testing.T) {
	checkAnywhere(t, `<?php
function e(string $v, int|string $w) {
    if (is_numeric($v)) { t('numeric', $v); return $v; }
    t('rest', $v);
    if (!is_numeric($w)) { t('notNumeric', $w); }
}
`, map[string]string{"numeric": "string", "rest": "string", "notNumeric": "string"})
}

// The casting typer gives up on variables also defined by constructs it
// does not type (destructuring, out arguments…).
func TestTRulesUnmodelledDefinitions(t *testing.T) {
	src := `<?php
function f(string $t, array $xs, $c) {
    list($h, $m) = explode(':', $t);
    if (!$h) { $h = 0; }
    t('list', $h);
    $p = 1;
    preg_match('/x/', $t, $p);
    t('out', $p);
    $q = 1;
    if ($c) { $q = 2; }
    t('plain', $q);
    /** @var int $d */
    $d = 1;
    t('doc', $d);
    foreach ($xs as $x) { t('binding', $x); }
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	want := map[string]string{"'list'": "?unknown", "'out'": "?unknown", "'plain'": "int", "'doc'": "int", "'binding'": "?unknown"}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != want[lit.Raw] {
					t.Errorf("%s: got %s want %s", lit.Raw, got, want[lit.Raw])
				}
			}
		}
		return true
	})
}

// A property read right after storing an unknown value is unknown unless
// its native type is enforced.
func TestPropertyAfterUnknownWrite(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
class C {
    /** @var string */
    protected $p;
    private ?Foo $n = null;
    public $pub;
    /** @var int */
    private static $s = 0;
    public function h($o, Foo $f) {
        $this->p = isset($o['k']) ? $o['k'] : null;
        t('doc', $this->p);
        $this->n = $o;
        t('native', $this->n);
        $this->p = 'x';
        t('known', $this->p);
        $this->p .= 'y';
        t('compound', $this->p);
        self::$s = $o;
        t('static', self::$s);
        $this->{'dyn'} = $o;
        $x = new C();
        $x->pub = $o;
        t('other', $x->pub);
    }
}
`, map[string]string{"doc": "?unknown", "native": `\Foo|null`, "known": "string", "compound": "string", "static": "?unknown", "other": "?unknown"})
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

// Properties extensions create but the stubs omit are declared.
func TestStubMissingProps(t *testing.T) {
	ix := stubs.Index()
	if ix.FindProperty("OAuthProvider", "nonce", phpversion.PHP84) == nil || ix.FindProperty("OAuthProvider", "bogus", phpversion.PHP84) != nil {
		t.Error("OAuthProvider properties")
	}
}
