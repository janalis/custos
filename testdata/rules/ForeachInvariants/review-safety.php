<?php
function adjust(&$x) {}
// The generated value variable already exists: reported, no fix.
function nameTaken(array $rows) {
    $iValue = 'total';
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($rows); $i++) { echo $rows[$i]; }
    return $iValue;
}
function nameTakenParam(array $rows, $iValue) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($rows); $i++) { echo $rows[$i]; }
}
$nameTakenImport = function () use ($rows, $iValue) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($rows); $i++) { echo $rows[$i]; }
};
$nameFree = function () use ($rows) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($rows); $i++) { echo $rows[$i]; }
};
// Accesses in other scopes keep the container and the key.
function scopes(array $names) {
    $out = [];
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($names); $i++) {
        $out[] = function () use ($names, $i) { return strtoupper($names[$i]); };
        $out[] = static function () use ($names, $i) { return strtoupper($names[$i]); };
        $out[] = fn() => strtoupper($names[$i]);
        $out[] = fn($x) => strtoupper($names[$i]) . $x;
        $out[] = fn($i) => strtoupper($names[$i]);
        $out[] = fn($names) => strtoupper($names[$i]);
        if (!function_exists('inner')) {
            function inner($names, $i) { return strtoupper($names[$i]); }
        }
        $out[] = new class {
            function pick($names, $i) { return strtoupper($names[$i]); }
        };
    }
    return $out;
}
// Writes below the element keep the container.
function subWrites(array $a) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { $a[$i]['n']++; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { unset($a[$i]['tmp']); }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { sort($a[$i]['t']); }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { foreach ($a[$i] as &$v) { $v = 0; } unset($v); }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { list($a[$i]['x'], $y) = [9, 8]; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { $x = [&$a[$i]['k']]; $x[0] = 1; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { return $a[$i]['r']; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) {
        echo $a[$i]['n'], strlen($a[$i]['s']), $a[$i]['o']->p, $a[$i]['a'] . $a[$i]['b'][0];
        if ($a[$i]['f']) {}
        foreach ($a[$i] as $w) {}
        foreach ($a[$i]['list'] as $w) {}
    }
    return $a;
}
function staleSubWrites(array $a) {
    for ($i = 0; $i < count($a); $i++) { foreach ($a[$i] as &$v) { echo $a[$i][0]; } }
    for ($i = 0; $i < count($a); $i++) { foreach ($a[$i] as &$v) { $v = 0; } echo $a[$i][0]; }
    for ($i = 0; $i < count($a); $i++) { foreach ($a as &$w) { $w = 1; } echo $a[$i]; }
    for ($i = 0; $i < count($a); $i++) { $a[$i]['n']++; echo $a[$i]['n']; }
}
// Counter or limit shared with other code.
function byRefLimit(array $a, &$n) {
    $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
}
$importedLimit = function (array $a) use (&$n) {
    $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
};
function byRefHeaderLimit(array $a, &$n) {
    for ($i = 0, $n = count($a); $i < $n; $i++) { echo $a[$i]; }
}
function byRefCounter(array $a, &$i) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
}
function globalCounter(array $a) {
    global $pos;
    for ($pos = 0; $pos < count($a); $pos++) { echo $a[$pos]; }
}
function staticCounter(array $a) {
    static $pos;
    for ($pos = 0; $pos < count($a); $pos++) { echo $a[$pos]; }
}
function aliasOfCounter(array $a) {
    $r = &$i;
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
}
function counterAsAlias(array $a, $r) {
    $i = &$r;
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
}
function importedByNestedClosure(array $a) {
    $f = function () use (&$i) { return $i; };
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    return $f;
}
$importedCounter = function (array $a) use (&$i) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
};
// Variables read by name.
function compactLimit(array $a) {
    $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
    return compact('n');
}
function compactOther(array $a) {
    $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
    return compact('a');
}
function compactCounter(array $a) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    return compact('a', 'i');
}
function compactDynamic(array $a, $names) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    return compact($names);
}
function compactSpread(array $a, $names) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    return compact(...$names);
}
function definedVars(array $a) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    return get_defined_vars();
}
function variableVariable(array $a, $x) {
    for ($i = 0; $i < count($a); $i++) { echo $a[$i], $$x; }
}
function dynamicBefore(array $a, $x) {
    echo $$x, compact('i');
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
}
// Read on the next pass of an enclosing loop.
function nextPass(array $a, $m) {
    $i = 5;
    $out = '';
    while ($m-- > 0) {
        $out .= $i;
        for ($i = 0; $i < count($a); $i++) { $out .= $a[$i]; }
    }
    return $out;
}
function nextPassDynamic(array $a, $m) {
    while ($m-- > 0) {
        echo compact('i');
        for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    }
}
function nextPassRedefined(array $a, $m) {
    while ($m-- > 0) {
        for ($i = 0; $i < 3; $i++) {}
        <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    }
}
function outerHeader(array $a) {
    for ($i = 0; $i < 2; $i++) {
        for ($i = 0; $i < count($a); $i++) { echo $a[$i]; }
    }
}
// A limit computed ahead of a change of the container.
function limitBeforeChange(array $a, array $extra) {
    $n = count($a);
    $a = array_merge($a, $extra);
    for ($i = 0; $i < $n; $i++) { echo $a[$i]; }
}
function limitTouched(array $a) {
    $n = count($a);
    adjust($n);
    for ($i = 0; $i < $n; $i++) { echo $a[$i]; }
}
function limitAcrossPasses(array $a, $m) {
    $n = count($a);
    while ($m-- > 0) {
        for ($i = 0; $i < $n; $i++) { echo $a[$i]; }
        $a[] = 1;
    }
}
function limitPerPass(array $a, $m) {
    while ($m-- > 0) {
        $n = count($a);
        <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
        $a[] = 1;
    }
}
function limitSometimes(array $a, $m) {
    while ($m-- > 0) {
        if ($m > 5) {
            $n = count($a);
        }
        for ($i = 0; $i < $n; $i++) { echo $a[$i]; }
        $a[] = 1;
    }
}
function limitInCondition(array $a, $m) {
    if ($n = count($a)) {
        while ($m-- > 0) {
            for ($i = 0; $i < $n; $i++) { echo $a[$i]; }
            $a[] = 1;
        }
    }
}
function limitUnbraced(array $a, $x) {
    if ($x) $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
}
class Box {
    public $n;
    public $items = [];
    function stale() {
        $this->n = count($this->items);
        $this->items[] = 1;
        for ($i = 0; $i < $this->n; $i++) { echo $this->items[$i]; }
    }
}
// each() targets read after the loop: reported, no fix.
function eachKeyAfter(array $a) {
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($a)) { if ($v === 'x') { break; } }
    return $k;
}
function eachValueAfter(array $a) {
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($a)) { echo $k; }
    return $v;
}
function eachByRef(array $a, &$k) {
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($a)) { echo $v; }
}
function eachTwice(array $a, array $b) {
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($a)) { echo $v; }
    <error descr="Replace the each() loop with foreach.">while</error> (list($k, $v) = each($b)) { echo $k, $v; }
    [$k, $v] = [1, 2];
    echo $k, $v;
}
// An element captured by reference may be written through the alias and
// read again: reported, no fix (the value variable would be a stale copy).
function capturedElements(array $a, $k) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { $r = &$a[$i]; $r = strtoupper($r); echo $a[$i]; unset($r); }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { $box = [&$a[$i]['k'], function () { return [&$x]; }]; $box[0] = 1; echo $a[$i]['k']; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { $r = &$a[$k]; echo $a[$i]; }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) { echo $a[$i]; $r = &$a[$i]; }
}
