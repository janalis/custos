<?php
namespace Clock;

function time() { return 0; }

function due($spec, $clock) {
    $a = strtotime('now ');
    $b = strtotime($spec, $now);
    $c = strtotime($spec, microtime(true));
    $e = strtotime();
    $f = strtotime('+1 hour', time());
    $g = strtotime('+1 hour', (\time()));
    $i = $clock->strtotime('now');
    $j = strtotime('+1 hour', $clock->time());
    $k = strtotime('now', 1, 2);
    return [$a, $b, $c, $e, $f, $g, $h, $i, $j, $k];
}

$parse = strtotime(...);
