<?php
class Box {
    /** @var int */
    public $size;
    /** @var Box|null */
    public $inner;
}
/** @var Box $box */
/** @var Box[] $boxes */

$a = $opts['k'] ?? 5;
$b = $opts['k'] ?? null;
$c = $row->id ?? ($fallback . 'x');
$d = $map[$k] ?? NULL;
$e = $box->size ?? 0;
$f = $box->inner ?? null;
$g = $boxes[1]->size ?? null;
$h = $box ? $box->inner : -1;            // nullable with non-null fallback
$i = $mystery ? $mystery->size : null;   // unknown type
$j = array_key_exists($k, $map) ? $map[$k] : 0;
$k2 = isset($opts['k'], $opts['j']) ? $opts['k'] : 5;
$l = isset($opts['k']) ?: 5;

function pickA($in) {
    $out = $in['v'] ?? 'none';
    $n = abs($n);
    if (isset($in[$n])) { $n = $in[$n]; }      // previous value reads the target
    return $in['w'] ?? 'none';
}

function pickB($in) {
    if ($in) {
        $res = 1;
    } else { $res = $in ?? 2; }
    return $in ?? null;
}

$cb = function ($q) {
    return $q ?? null;
};
