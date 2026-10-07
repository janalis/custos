<?php
function f($rows) {
    foreach ($rows as $r) {
        foreach ($r as $c) {
            $a[] = $c;
            $b[] = $c;
            $d[] = $c;
            $e[] = $c;
            $g[] = $c;
            $h[] = $c;
            $i[] = $c;
            self::$cache[] = $c;
            f()[] = $c;
        }
    }
    $a = load();
    $copy = $b;
    $d .= '';
    foreach ($e as $v) {}
    foreach ($rows as $g) {}
    foreach ($rows as $h => $v) {}
    [$i] = $rows;
}

function g($rows) {
    global $k;
    static $s = [];
    foreach ($rows as $r) {
        for (;;) {
            $k[] = $r;
            $s[] = $r;
        }
    }
}

$fn = fn($rows) => array_map(function () { return 1; }, $rows);
