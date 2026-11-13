package infer_test

import "testing"

// A variable assigned in an operand of `&&` (or, negated, `||`) is assigned
// whenever the condition holds: the body of the if, elseif or while it
// guards does not read it as possibly undefined (null).
func TestConditionAssignsWhenTrue(t *testing.T) {
	checkAnywhere(t, `<?php
function f(?string $d, $c) {
    if (null !== $d && false !== ($r = realpath($d))) {
        t('and', $r);
    }
    if ($c) {
    } elseif ($c !== 1 and ($e = realpath('x'))) {
        t('elseif', $e);
    }
    while (!(!$c || !($w = realpath('y')))) {
        t('notOr', $w);
    }
    if ($c && ($a = new ArrayObject())->count() && t('andRight', $a)) {
    }
    if (!$c || !($b = realpath('b')) || t('orRight', $b)) {
    }
    if ($c || ($o = realpath('z'))) {
        t('or', $o);
    }
}
`, map[string]string{"and": "string", "elseif": "string", "notOr": "string", "andRight": `\ArrayObject`, "orRight": "string", "or": "false|null|string"})
}

// An if chain whose branches each assign a variable or leave the loop
// around it (continue, break) leaves the variable assigned after it.
func TestBranchesAssignOrJumpOut(t *testing.T) {
	checkAnywhere(t, `<?php
class Pkg {}
function make(): Pkg { return new Pkg(); }
function f(array $ops) {
    foreach ($ops as $op) {
        if ($op === 1) {
            $p = make();
        } elseif ($op === 2) {
            $p = make();
        } else {
            continue;
        }
        t('loop', $p);
    }
    switch (count($ops)) {
        case 1:
            if ($ops) {
                $q = make();
            } else {
                break;
            }
            t('switch', $q);
    }
    foreach ($ops as $op) {
        if ($op) {
            $r = make();
        } else {
            break;
        }
    }
    t('afterLoop', $r);
}
`, map[string]string{"loop": `\Pkg`, "switch": `\Pkg`, "afterLoop": `\Pkg|null`})
}
