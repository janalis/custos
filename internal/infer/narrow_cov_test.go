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

// checkAnywhere is checkWith (ShapeString) for every `t('label', expr)`
// call, including those that are not expression statements.
func checkAnywhere(t *testing.T, src string, want map[string]string) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && len(c.Args.Args) == 2 {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).ShapeString()
			}
		}
		return true
	})
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %q want %q", k, got[k], w)
		}
	}
}

func TestNarrowStructure(t *testing.T) {
	checkAnywhere(t, `<?php
/** @param int|string|null $x */
function f($x, $c) {
    if ($c) {} elseif (is_int($x)) { t('elseif', $x); }
    switch ($c) {
        case 1:
            if (!is_string($x)) { return; }
            t('case', $x);
            break;
        case t('caseCond', $x):
            break;
    }
    if (is_int($x)) { $g = fn() => t('arrow', $x); }
}
`, map[string]string{
		"elseif":   "int",
		"case":     "string",
		"caseCond": "int|null|string", // a case expression: no guard applies
		"arrow":    "int",             // the arrow captures the narrowed value at construction
	})
}

func TestNarrowConditions(t *testing.T) {
	checkWith(t, nil, `<?php
/**
 * @param int|string|null $x
 * @param int|null $n
 * @param true|mixed $sm
 * @param int|string $is
 * @param int[]|null $a
 */
function f($x, $n, $sm, $is, $a, $cls, $fn) {
    if (!(is_int($x) || is_string($x))) { t('notOr', $x); }
    if (!($x === null)) { t('notIdentNull', $x); }
    if ($x === 'a') { t('identStr', $x); }
    if ($x === PHP_EOL) { t('identConst', $x); }
    if ($n === false) { t('identFalseAbsent', $n); }
    if (!($a === [])) { t('notIdentEmpty', $a); }
    if ($a === []) { t('identEmpty', $a); }
    if ($a != []) { t('neEmpty', $a); }
    if (!($a != [])) { t('notNeEmpty', $a); }
    if (count($a) == 0) { t('countEq0', $a); }
    if (count($a) == 2) { t('countEq2', $a); }
    if ($x instanceof $cls) { t('instDyn', $x); }
    if ($fn($x)) { t('dynCall', $x); }
    if (rand()) { t('noArgs', $x); }
    if (count($a)) { t('count', $a); }
    if (is_bool($sm)) { t('boolMixed', $sm); }
    if (is_bool($is)) { t('boolNone', $is); }
    if (!is_int($n) && !is_null($n)) { t('contradiction', $n); }
    if ($fn($a) > 0) { t('dynCount', $a); }
    if (strlen($a) > 0) { t('notCount', $a); }
    if (count($a) > $n) { t('countVar', $a); }
    if (count($a) > '1') { t('countStr', $a); }
    if (0 < count($a)) { t('flipLess', $a); }
    if (1 > count($a)) { t('flipGreater', $a); }
    if (1 <= count($a)) { t('flipLe', $a); }
    if (0 >= count($a)) { t('flipGe', $a); }
    if (!(count($a) < 1)) { t('notLess', $a); }
    if (!(count($a) > 0)) { t('notGreater', $a); }
    if (!(count($a) <= 0)) { t('notLe', $a); }
    if (!(count($a) >= 1)) { t('notGe', $a); }
    if (!(count($a) !== 0)) { t('notNotIdent', $a); }
    if (count($a) >= 2) { t('ge2', $a); }
    if (count($a) < 5) { t('less5', $a); }
}
`, map[string]string{
		"notOr":            "null",
		"notIdentNull":     "int|string",
		"identStr":         "string",
		"identConst":       "string",
		"identFalseAbsent": "int|null",
		"notIdentEmpty":    "non-empty int[]|null",
		"identEmpty":       "int[]|null",
		"neEmpty":          "non-empty int[]",
		"notNeEmpty":       "int[]|null",
		"countEq0":         "int[]|null",
		"countEq2":         "non-empty int[]",
		"instDyn":          "int|null|string",
		"dynCall":          "int|null|string",
		"noArgs":           "int|null|string",
		"count":            "non-empty int[]",
		"boolMixed":        "bool|true",
		"boolNone":         "bool",
		"contradiction":    "?unknown",
		"dynCount":         "int[]|null",
		"notCount":         "int[]|null",
		"countVar":         "int[]|null",
		"countStr":         "int[]|null",
		"flipLess":         "non-empty int[]",
		"flipGreater":      "int[]|null",
		"flipLe":           "non-empty int[]",
		"flipGe":           "int[]|null",
		"notLess":          "non-empty int[]",
		"notGreater":       "int[]|null",
		"notLe":            "non-empty int[]",
		"notGe":            "int[]|null",
		"notNotIdent":      "int[]|null",
		"ge2":              "non-empty int[]",
		"less5":            "int[]|null",
	})
}

func TestNarrowKeyConditions(t *testing.T) {
	checkWith(t, nil, `<?php
/** @param array{a?: int|null, b?: string} $s */
function f(array $s, $z) {
    if (!isset($s['a']) || $z) { return; }
    t('orGuard', $s);
    if (in_array('x', $s)) { t('otherCall', $s); }
    if (array_key_exists(key: 'b', array: $s)) { t('named', $s); }
}
/** @param int|false $x */
function g($x) {
    if (false === $x) { $x .= 'a'; }
    t('compound', $x);
}
`, map[string]string{
		"orGuard":   "array{a: int, b?: string}",
		"otherCall": "array{a: int, b?: string}",
		"named":     "array{a: int, b?: string}",
		"compound":  "false|int|string",
	})
}
