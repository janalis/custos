package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// Edge cases of comparison and type-guard narrowing: shapes the guards do
// not understand leave the type alone.
func TestNarrowCompleteEdges(t *testing.T) {
	checkAnywhere(t, `<?php
const ANSWER = 42;
class K { const S = 'a'; public ?K $n = null; public int $i = 0; }
/** @phpstan-assert string $v */
function assertString($v) {}
/**
 * @param int|string|null $x
 * @param K|string|null $o
 */
function f($x, $o, K $kk, null $nul, ?K $k, ?K $k2, ?K $k3, ?K $k4, mixed $m, $f, $y, $cls) {
    if ($nul?->n) { t('nullBase', $nul); }
    if ($k?->i === ANSWER) { t('constChain', $k); }
    if ($k2?->i === -1) { t('negChain', $k2); }
    if ($k3?->i === $y) { t('varChain', $k3); }
    if ($k4?->i == 1) { t('looseChain', $k4); }
    if ($x === -ANSWER) { t('negConst', $x); }
    if ($x === Unknown::FOO) { t('unknownConst', $x); }
    if ($x === $y) { t('identVar', $x); }
    if ($x === K::S) { t('classConst', $x); }
    if (true !== $f($x)) { t('dynBool', $x); }
    if (true !== in_array($x, [1], true)) { t('inArrayBool', $x); }
    if (true !== $y) { t('otherBool', $x); }
    if (true === true) { t('constConst', $x); }
    if (strlen($x) === 3) { t('strlenCmp', $x); }
    if (gettype($x, 1) === 'string') { t('gettypeArgs', $x); }
    if (gettype(...$x) === 'string') { t('gettypeUnpack', $x); }
    if ($f($x) === 'string') { t('dynFunc', $x); }
    if ($o::FOO === 1) { t('notClassConst', $o); }
    if ($y::class === K::class) { t('otherClass', $o); }
    if (get_class($o) === $y) { t('getClassVar', $o); }
    if (get_class($o) === K::S) { t('getClassConst', $o); }
    if (get_class($o) === $cls::class) { t('getClassDyn', $o); }
    if (get_class($o) === 1) { t('getClassInt', $o); }
    if (get_class($o) === 'a b') { t('getClassBad', $o); }
    if (get_class($o) === 'A\\\\B') { t('getClassEmptyPart', $o); }
    if (get_class($o) === '1K') { t('getClassDigit', $o); }
    if (get_class($o) === '\K') { t('getClassLeading', $o); }
    if (is_a($o)) { t('isAOne', $o); }
    if (is_a($o, class: K::class)) { t('isANamed', $o); }
    if (is_a($o, $cls)) { t('isADyn', $o); }
    if (is_a($o, K::class, $y)) { t('isADynString', $o); }
    if (in_array($x, ['a'])) { t('inArrayTwo', $x); }
    if (in_array($x, ['a'], strict: true)) { t('inArrayNamed', $x); }
    if (in_array($x, ['a'], false)) { t('inArrayFalse', $x); }
    if (in_array($x, [], true)) { t('inArrayEmpty', $x); }
    if (in_array($x, [$m], true)) { t('inArrayMixed', $x); }
    if (in_array($x, $f, true)) { t('inArrayUnknown', $x); }
    if (trim($x)) { t('otherFunc', $x); }
    if (!($x === null && $x === null)) { t('emptyRight', $x); }
    if (!($kk instanceof K && $y)) { t('emptyLeft', $kk); }
    if (isset($a['k'])) { return; }
}
/** @param int|string|null $x */
function g($x, ?string $s) {
    if ($x === null) { return; }
    if ($x === 1) { return; }
    $x = null;
    if ($x === null) { return; }
    assertString($x);
    t('assertAfterEmpty', $x);
    if (isset($s['k'])) { return; }
    t('issetFalse', $s);
}
`, map[string]string{
		"nullBase":          "null",
		"constChain":        `\K`,
		"negChain":          `\K`,
		"varChain":          `\K|null`,
		"looseChain":        `\K|null`,
		"negConst":          "int|null|string",
		"unknownConst":      "int|null|string",
		"identVar":          "int|null|string",
		"classConst":        "string",
		"dynBool":           "int|null|string",
		"inArrayBool":       "int|null|string",
		"otherBool":         "int|null|string",
		"constConst":        "int|null|string",
		"strlenCmp":         "int|null|string",
		"gettypeArgs":       "int|null|string",
		"gettypeUnpack":     "int|null|string",
		"dynFunc":           "int|null|string",
		"notClassConst":     `\K|null|string`,
		"otherClass":        `\K|null|string`,
		"getClassVar":       `\K|null|string`,
		"getClassConst":     `\K|null|string`,
		"getClassDyn":       `\K|null|string`,
		"getClassInt":       `\K|null|string`,
		"getClassBad":       `\K|null|string`,
		"getClassEmptyPart": `\K|null|string`,
		"getClassDigit":     `\K|null|string`,
		"getClassLeading":   `\K`,
		"isAOne":            `\K|null|string`,
		"isANamed":          `\K|null|string`,
		"isADyn":            `\K|null|string`,
		"isADynString":      `\K|string`,
		"inArrayTwo":        "int|null|string",
		"inArrayNamed":      "int|null|string",
		"inArrayFalse":      "int|null|string",
		"inArrayEmpty":      "int|null|string",
		"inArrayMixed":      "int|null|string",
		"inArrayUnknown":    "int|null|string",
		"otherFunc":         "int|null|string",
		"emptyRight":        "int|string",
		"emptyLeft":         `\K`,
		"assertAfterEmpty":  "?unknown",
		"issetFalse":        "null|string",
	})
}

// Branch conditions: definitions between the conditions and the read,
// several values in one arm, wide matches and switches.
func TestNarrowBranchEdges(t *testing.T) {
	var wide strings.Builder
	wide.WriteString("<?php\nfunction w(int|string|null $x) {\n $a = match (true) {\n")
	for i := 0; i < 300; i++ {
		wide.WriteString("  $x === 1 => 1,\n")
	}
	wide.WriteString("  is_string($x), is_int($x) => t('late', $x),\n  default => t('wideDefault', $x),\n };\n switch (true) {\n")
	for i := 0; i < 300; i++ {
		wide.WriteString("  case $x === 1:\n")
	}
	wide.WriteString("  case is_string($x): t('wideGroup', $x); break;\n }\n}\n")
	checkAnywhere(t, wide.String(), map[string]string{
		"late":        "int|string",
		"wideDefault": "int|null|string",
		"wideGroup":   "int|null|string",
	})
	checkAnywhere(t, `<?php
function f(int|string|null $x, int $c) {
    $a = match (true) {
        is_int($x) => 1,
        default => ($x = $c) . t('defAfterDef', $x),
    };
    $b = match (true) {
        is_int($x), t('secondCond', $x) => 1,
        default => 2,
    };
    $c2 = match (true) {
        is_int($x) => ($x = $c) . t('armAfterDef', $x),
        default => 2,
    };
    switch (true) {
        case is_int($x):
            break;
        case is_string($x):
            $x = $c;
            t('caseAfterDef', $x);
            break;
        default:
            $x = $c;
            t('defaultAfterDef', $x);
    }
    switch ($c) {
        case 1: { $x = 1; break; }
        case 2: t('afterBlockBreak', $x); break;
        case t('caseCond', $x): break;
    }
}
`, map[string]string{
		"defAfterDef":     "int|null|string",
		"secondCond":      "null|string",
		"armAfterDef":     "int|null|string",
		"caseAfterDef":    "int|null|string",
		"defaultAfterDef": "int|null|string",
		"afterBlockBreak": "int|null|string",
		"caseCond":        "int|null|string",
	})
}

// A branch missing from its parent's list (only possible in a mutated
// tree) narrows by nothing but its own condition.
func TestNarrowBranchNotListed(t *testing.T) {
	src := `<?php
function f(int|string|null $x, $c) {
    if (is_int($x)) {} elseif ($x === null) { t('elseif', $x); } else { t('else', $x); }
    $a = match (true) { is_int($x) => 1, is_string($x) => t('arm', $x), default => 2 };
    switch (true) { case is_int($x): break; case is_string($x): t('case', $x); break; }
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	var calls []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				calls = append(calls, c)
			}
		}
		return true
	})
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.If:
			n.ElseIfs = nil
		case *syntax.Match:
			n.Arms = n.Arms[:1]
		case *syntax.Switch:
			n.Cases = n.Cases[:1]
		}
		return true
	})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	got := map[string]string{}
	for _, c := range calls {
		label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
		got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
	}
	want := map[string]string{"elseif": "null", "arm": "int|null|string", "case": "int|null|string"}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %s want %s", k, got[k], w)
		}
	}
}

// Plain expression kinds whose coverage used to come only from rule
// fixtures that narrowing now types differently.
func TestPlainExpressionKinds(t *testing.T) {
	checkAnywhere(t, `<?php
class SP { public static int $n = 0; }
function f(int $i, $u, string $s) {
    t('staticUnknownClass', $u::$n);
    t('staticMissing', SP::$missing);
    t('static', SP::$n);
    t('file', __FILE__);
    t('floatCast', (float) $u);
    t('divUnknown', $u / $i);
    t('div', $i / 2);
    t('noClass', $i->m());
    t('paren', ($s));
}
`, map[string]string{
		"staticUnknownClass": "?unknown",
		"staticMissing":      "?unknown",
		"static":             "int",
		"file":               "string",
		"floatCast":          "float",
		"divUnknown":         "?unknown",
		"div":                "float|int",
		"noClass":            "?unknown",
		"paren":              "string",
	})
}
