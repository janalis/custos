<?php
function &byRef() {
    $slot = null;
    return $slot;
}

function twice() {
    $n = compute();
    $n = $n * 2;
    return $n;
}

function readAgain() {
    $cfg = load();
    [$k] = $cfg;
    return $cfg ? $k : null;
}

function annotated() {
    /** @var Gadget $g */
    $g = factory();
    return $g;
}

function inlineAnnotated() {
    /* @var Gadget $g */
    $g = factory();
    return $g;
}

function staticAccess() {
    $cls = whichClass();
    return $cls::build();
}

function compound($v) {
    $v .= '!';
    return $v;
}

$fn = function () use (&$acc) {
    $acc = 5;
    return $acc;
};

function refParam(&$out) {
    $out = 1;
    return $out;
}

function separated() {
    $x = 1;
    /** doc between */
    return $x;
}

function chained() {
    $o = make();
    return $o->a->b;
}

function element() {
    $a = make();
    return $a[0];
}

function expression() {
    $a = make();
    return -$a;
}

function otherVar() {
    $a = make();
    return $b;
}

function nestedDestructuring() {
    $a = make();
    foreach ([$a] as [$x]) {}
    return $x;
}

function longOne() {
    $value = someVeryLongFunctionName($firstArgument, $secondArgument, $thirdArgument, 42);
    return $value;
}

function usedInClosure() {
    $v = make();
    $f = function () use ($v) { return $v; };
    return $v;
}

function namelessAnnotation($auth) {
    /** @var \App\Guard */
    $guard = $auth->guard('api');
    return $guard;
}

function refDestructuring() {
    $pair = [1, 2];
    [&$a, $b] = $pair;
    $a = 9;
    return $a + $b;
}

function nestedRefDestructuring() {
    $rows = load();
    list($first, list(&$cell)) = $rows;
    return $cell;
}

class Slots {
    public $items = [];
    public function slot($k) {
        $v = &$this->items[$k];
        return $v;
    }
}

function compacted() {
    $parts = explode(':', 'a:b');
    [$x, $y] = $parts;
    return $x . $y . count(compact('parts'));
}

function variableVariable() {
    $parts = explode(':', 'a:b');
    [$x, $y] = $parts;
    $n = 'parts';
    return $x . $y . count($$n);
}

function definedVars() {
    try {
        $res = strtoupper('abc');
        return $res;
    } finally {
        audit(get_defined_vars());
    }
}

function template() {
    $data = load();
    [$title, $body] = $data;
    include 'view.php';
}

function refClosure() {
    return function &() {
        $slot = 5;
        return $slot;
    };
}
