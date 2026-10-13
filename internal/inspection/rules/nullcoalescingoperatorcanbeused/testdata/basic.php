<?php
class Box {
    /** @var int */
    public $size;
    /** @var Box|null */
    public $inner;
}
/** @var Box $box */
/** @var Box[] $boxes */

$a = <weak_warning descr="Simplify to '$opts['k'] ?? 5' using the null coalescing operator.">isset($opts['k']) ? $opts['k'] : 5</weak_warning>;
$b = <weak_warning descr="Simplify to '$opts['k'] ?? null' using the null coalescing operator.">!isset($opts['k']) ? null : $opts['k']</weak_warning>;
$c = <weak_warning descr="Simplify to '$row->id ?? ($fallback . 'x')' using the null coalescing operator.">null === $row->id ? $fallback . 'x' : $row->id</weak_warning>;
$d = <weak_warning descr="Simplify to '$map[$k] ?? NULL' using the null coalescing operator.">(array_key_exists($k, $map)) ? $map[$k] : NULL</weak_warning>;
$e = <weak_warning descr="Simplify to '$box->size ?? 0' using the null coalescing operator.">$box ? $box->size : 0</weak_warning>;
$f = <weak_warning descr="Simplify to '$box->inner ?? null' using the null coalescing operator.">!$box ? null : $box->inner</weak_warning>;
$g = <weak_warning descr="Simplify to '$boxes[1]->size ?? null' using the null coalescing operator.">!empty($boxes[1]) ? $boxes[1]->size : null</weak_warning>;
$h = $box ? $box->inner : -1;            // nullable with non-null fallback
$i = $mystery ? $mystery->size : null;   // unknown type
$j = array_key_exists($k, $map) ? $map[$k] : 0;
$k2 = isset($opts['k'], $opts['j']) ? $opts['k'] : 5;
$l = isset($opts['k']) ?: 5;

function pickA($in) {
    $out = 'none';
    <weak_warning descr="Simplify to '$out = $in['v'] ?? 'none'' using the null coalescing operator.">if</weak_warning> (isset($in['v'])) {
        $out = $in['v'];
    }
    $n = abs($n);
    if (isset($in[$n])) { $n = $in[$n]; }      // previous value reads the target
    <weak_warning descr="Simplify to 'return $in['w'] ?? 'none'' using the null coalescing operator.">if</weak_warning> ($in['w'] !== null) {
        return $in['w'];
    }
    return 'none';
}

function pickB($in) {
    if ($in) {
        $res = 1;
    } else <weak_warning descr="Simplify to '$res = $in ?? 2' using the null coalescing operator.">if</weak_warning> (isset($in)) {
        $res = $in;
    } else {
        $res = 2;
    }
    <weak_warning descr="Simplify to 'return $in ?? null' using the null coalescing operator.">if</weak_warning> (!(null === $in)) {
        return $in;
    } else {
        return null;
    }
}

$cb = function ($q) {
    <weak_warning descr="Simplify to 'return $q ?? null' using the null coalescing operator.">if</weak_warning> (isset($q)) {
        return $q;
    }
};
