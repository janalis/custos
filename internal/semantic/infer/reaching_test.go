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

// The back edge of a do-while (or for step) re-enters the body only when
// the condition held: `while ($e = $e->getPrevious())` loops with a truthy
// $e.
func TestDoWhileBackEdge(t *testing.T) {
	checkAnywhere(t, `<?php
function f(\Exception $e, ?array $rows, $c) {
    $all = [];
    do {
        $all[] = $e;
        t('body', $e);
    } while ($e = $e->getPrevious());
    t('all', $all);
    t('after', $e);
    for ($r = $rows; $r !== null; $r = $c ? null : $rows) {
        t('for', $r);
    }
    do { t('plain', $c); $c = rand() ? null : 1; } while ($c);
    for ($n = $rows; $n !== null; $n = null) { t('nullStep', $n); }
}
`, map[string]string{
		"body": `\Exception|\Throwable`, "all": `\Exception[]|\Throwable[]`, "after": `\Exception|\Throwable|null`, // the pre-loop value is not known to be overwritten
		"for": "array", "plain": "?unknown", "nullStep": "array",
	})
}

// A then-branch reassignment does not reach the else branch, so the
// condition still narrows there; definitions on paths that always leave
// (return, throw, exit, break out of the region) reach nothing after.
func TestExclusiveBranchesAndExits(t *testing.T) {
	checkAnywhere(t, `<?php
function f(array|int $c, $cc, int $i) {
    if (is_int($c)) { $c = []; } else { t('else', $c); }
    $v = $cc ? ($i = 'a') : t('ternary', $i);
    $m = match (true) { $cc === 1 => $i = 2.5, default => t('arm', $i) };
    while ($cc) {
        if (is_int($c)) { $c = 'x'; } else { t('loopElse', $c); }
    }
    $x = 5;
    if ($cc) { $x = 'a'; return 1; }
    t('afterReturn', $x);
    $y = 1;
    if ($cc) { $y = 'b'; throw new \Exception(); }
    t('afterThrow', $y);
    $z = 1;
    if ($cc) { $z = 'c'; } else { $z = 2.5; exit; }
    t('afterIfElse', $z);
    try {
        $t = 1;
        if ($cc) { $t = 'd'; throw new \Exception(); }
    } catch (\Exception $e) { t('catch', $t); }
    foreach ([1] as $k) {
        $w = 1;
        if ($cc) { $w = 'e'; break; }
        t('inLoop', $w);
    }
    t('afterBreak', $w);
    foreach ([1] as $k) {
        foreach ([2] as $l) {
            if ($cc) { $q = 'f'; break; }
        }
        t('outerLoop', $q);
        $q = 1;
    }
    foreach ([1] as $k) {
        $p = 1;
        if ($cc) { $p = 'g'; continue; }
        t('continued', $p);
    }
    $o = 1;
    foreach ([1] as $k) {
        t('nextIteration', $o);
        if ($cc) { $o = 'j'; continue; }
        if ($k) { $o2 = 2.5; continue; }
        t('viaHead', $o);
    }
    switch ($cc) {
        case 1:
            switch ($i) {
                case 2:
                    $n = 'k';
                    break;
                default:
                    t('innerSwitch', $n ?? 0);
            }
            t('afterInner', $n ?? 0);
    }
    for ($r = 0; ; $r = 'l') { t('noCond', $r); break; }
    switch ($cc) {
        case 1:
            $s = 'h';
            break;
        default:
            t('switchDefault', $s ?? null);
    }
    $u = 1;
    if ($cc) { $u = 'i'; return; t('dead', $u); }
}
`, map[string]string{
		"else": "array", "ternary": "int", "arm": "int|string", "loopElse": "array|int|string",
		"afterReturn": "int", "afterThrow": "int", "afterIfElse": "string", "catch": "int|null|string",
		"inLoop": "int", "afterBreak": "int|null|string", "outerLoop": "int|null|string", "continued": "int",
		"nextIteration": "int|string", "viaHead": "int|string", "innerSwitch": "?unknown", "afterInner": "int|string", "noCond": "int|string",
		"switchDefault": "?unknown", "dead": "string",
	})
}

// A variable that only plain assignments define, read on a path that
// assigns it nowhere, may be null.
func TestPossiblyUndefined(t *testing.T) {
	checkAnywhere(t, `<?php
function f($t, $c, array $xs) {
    if ($t) { $r = 'X'; }
    t('maybe', $r);
    if ($t) { $a = 1; } else { $a = 2; }
    t('ifElse', $a);
    switch ($c) { case 1: $s = 'a'; break; default: $s = 'b'; }
    t('switchDefault', $s);
    switch ($c) { case 1: $s2 = 'a'; break; case 2: $s2 = 'b'; break; }
    t('switchNoDefault', $s2);
    try { $u = 1; } catch (\Exception $e) { $u = 2; }
    t('tryCatch', $u);
    do { $d = 1; } while ($c);
    t('doWhile', $d);
    if (($w = strlen('x')) > 0) {}
    t('condAssign', $w);
    if ($t) { $z = 1; } else { return; }
    t('elseReturns', $z);
    foreach ($xs as $x) { $l = 1; }
    t('afterLoop', $l);
    $c2 = $c && ($v = 1);
    t('rhsAssign', $v);
    if (isset($r)) { t('isset', $r); }
    while (($line = fgets(STDIN)) !== false) { t('whileCond', $line); }
    for ($i = 0; $i < 3; $i++) { t('forInit', $i); }
    echo $e2 = 'x';
    t('echo', $e2);
    $tern = $t ? ($q = 1) : 2;
    t('ternary', $q);
    if ($t) { $k = 1; } elseif ($c) { $k = 2; } else { $k = 3; }
    t('chain', $k);
    if ($t) { $k2 = 1; } elseif ($c) { } else { $k2 = 3; }
    t('chainGap', $k2);
    try { $f1 = 1; } finally { $f2 = 2; }
    t('finally', $f2);
    t('tryBody', $f1);
    try { $g1 = 1; } catch (\Exception $e) { }
    t('catchGap', $g1);
    { $blk = 1; }
    t('block', $blk);
    while ($c) { $wb = 1; }
    t('whileBody', $wb);
    foreach ($xs as $fx) { }
    $fe = null;
    foreach (($fe2 = $xs) as $y) {}
    t('foreachExpr', $fe2);
    switch ($sw = $c) { default: }
    t('switchCond', $sw);
    $m = match ($mm = $c) { default => 1 };
    t('matchCond', $mm);
    if ($t) { $br = 1; } else { $br = 2; }
    switch ($c) { case 1: $sc = 1; case 2: $sc = 2; break; default: $sc = 3; }
    t('fallthrough', $sc);
    switch ($c) { case 1: $se = 1; break; default: }
    t('emptyDefault', $se);
    foreach ($xs as $x) { if ($x) { $cb = 1; continue; } $cb = 2; }
    t('continueList', $cb);
}
`, map[string]string{
		"maybe": "null|string", "ifElse": "int", "switchDefault": "string", "switchNoDefault": "null|string",
		"tryCatch": "int", "doWhile": "int", "condAssign": "int", "elseReturns": "int", "afterLoop": "int|null",
		"rhsAssign": "int|null", "isset": "string", "whileCond": "string", "forInit": "float|int", "echo": "string",
		"ternary": "int|null", "chain": "int", "chainGap": "int|null", "finally": "int", "tryBody": "int",
		"catchGap": "int|null", "block": "int", "whileBody": "int|null", "foreachExpr": "array", "switchCond": "?unknown",
		"matchCond": "?unknown", "fallthrough": "int", "emptyDefault": "int|null", "continueList": "int|null",
	})
}

// Conditions evaluated before a read assign the variable on every path.
func TestPossiblyUndefinedConditions(t *testing.T) {
	checkAnywhere(t, `<?php
function f($t, $c, array $xs) {
    if ($t) {} elseif ($e = $c) { t('elseifCond', $e); }
    foreach (($fx = $xs) as $y) { t('foreachExprIn', $fx); }
    switch ($s = $c) { case 1: t('switchIn', $s); }
    $m = match ($mm = $c) { 1 => t('matchIn', $mm), default => 0 };
    $b = ($bb = $c) && t('rhs', $bb);
    $q = ($qq = $c) ? t('ternaryIn', $qq) : 0;
}
`, map[string]string{"elseifCond": "?unknown", "foreachExprIn": "array", "switchIn": "?unknown", "matchIn": "?unknown", "rhs": "?unknown", "ternaryIn": "?unknown"})
	checkAnywhere(t, `<?php
function g(int $c, array $xs) {
    if ($c) {} elseif ($e = $c) { t('elseifCond', $e); }
    switch ($s = $c) { case 1: t('switchIn', $s); }
    $m = match ($mm = $c) { 1 => t('matchIn', $mm), default => 0 };
    $b = ($bb = $c) && t('rhs', $bb);
    $q = ($qq = $c) ? t('ternaryIn', $qq) : 0;
    for ($fi = 0; $fi < 3; $fi++) {}
    t('afterFor', $fi);
    if ($c) { $kc = 1; }
    switch ($c) { case 1: $kc2 = 1; break; case t('caseCond', $kc): break; }
}
/**
 * @param object|\ArrayIterator<int, string> $o
 * @param string|\ArrayIterator $x
 */
function h($o, $x) {
    if ($o instanceof \ArrayIterator) { t('args', $o); }
    if ($x instanceof \Countable) { return; } elseif ($x === '') { throw new \Exception(); }
    t('leaves', $x);
}
`, map[string]string{
		"elseifCond": "int", "switchIn": "int", "matchIn": "int", "rhs": "int", "ternaryIn": "int",
		"args": `\ArrayIterator<int, string>`, "leaves": "string", "afterFor": "float|int", "caseCond": "int|null",
	})
	src := `<?php
function k($x) {
    t('later', $u);
    $u = 1;
    t('pr', print_r($x, true));
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	for _, spec := range []bool{false, true} {
		tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
		tr.SpecOnly = spec
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
					lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
					got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
					if want := map[string]string{"'later'": "?unknown", "'pr'": "string"}[lit.Raw]; got != want {
						t.Errorf("spec=%v %s: %s want %s", spec, lit.Raw, got, want)
					}
				}
			}
			return true
		})
	}
}

// A definition always run earlier in the enclosing condition hides the
// definitions made before it (`preg_match(…, $m) && $m[1]`).
func TestDominatingConditionDefinitions(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $s, $c) {
    preg_match_all('/(a)/', $s, $m);
    if (preg_match('/(a)/', $s, $m, 0) && t('and', $m)) {}
    $x = 1;
    if (($x = 'a') && $c) { t('ifBody', $x); }
    $y = 1;
    while ($c && ($y = 'b')) { t('andCond', $y); }
    $y2 = 1;
    if ($c || ($y2 = 'b')) { t('conditional', $y2); }
    $n = 1;
    if (!($c || ($n = 'x'))) { t('notOr', $n); }
    $o = 1;
    $r = ($c || ($o = 'x')) || t('orFalse', $o);
    $z = 1;
    if ($c ?: ($z = 'c')) { t('ternary', $z); }
    $w = 1;
    $v = ($w = 'd') || t('or', $w);
    $a2 = 1;
    if (($a2 = 'e') || $c) { t('orLeft', $a2); }
    $q2 = 1;
    if (strlen($c ?? ($q2 = 'f'))) { t('coalesceArg', $q2); }
    $m2 = 1;
    if (match ($c) { 1 => ($m2 = 'g'), default => 0 }) { t('matchArm', $m2); }
}
`, map[string]string{
		"and": "string[]", "ifBody": "string", "andCond": "string", "conditional": "int|string", "notOr": "string", "orFalse": "string", "orLeft": "string", "coalesceArg": "int|string", "matchArm": "int|string", "ternary": "int|string", "or": "string",
	})
}
