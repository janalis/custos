<?php
function probe($n, $m, $a, $b, $obj) {
    $r = $n !== 7;
    $r = !($n <= 7);
    $r = (bool)($m % 2);
    $r = !($m | 4);
    $r = ($a || $b);
    $r = !($a || $b);
    $r = (bool)($obj instanceof Countable);
    $r = $n == $m;
    $r = $n == $m;
    $r = !($a and $b);
    $r = ($a < $b);
    return $r;
}
