<?php
function peek(string $buf, int $at, array $lines, $any)
{
    $d = mb_substr($buf, $at, 1);
    $e = substr($buf, $at, 2);
    $f = substr(strrev($buf), 0, 1);
    $g = substr($at, 0, 1);
    $h = substr($any, 0, 1);
    $i = substr($buf, 0);
    $j = substr($buf, 0, 01);
    $k = substr($buf, 0, -1);
    $m = substr((string) $at, 0, 1);
    $n = substr($lines, 0, 1);
    $o = substr(($buf), 0, 1);
    return [$d, $e, $f, $g, $h, $i, $j, $k, $l, $m, $n, $o];
}
