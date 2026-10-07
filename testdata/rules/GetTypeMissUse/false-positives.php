<?php
function f($v) {
    $a = gettype($v) < 'string';
    $b = gettype($v, 1) === 'string';
    $c = App\gettype($v) === 'string';
    $d = gettype($v) . 'x' === 'string';
    $e = gettype($v) === getKind();
    return [$a, $b, $c, $d, $e];
}
$x = gettype($y) === $kind;
