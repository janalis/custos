<?php
function next_turn(string $a, string $b, $n) {
    $v = glob($a);
    if ($v === false) { return; }
    while ($n--) {
        foreach (<warning descr="Handle glob failure before iterating its result.">$v</warning> as $entry) {}
        $v = glob($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = glob($a);
    if ($v === false) { return; }
again:
    foreach (<warning descr="Handle glob failure before iterating its result.">$v</warning> as $entry) {}
    $v = glob($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = glob($a);
    while ($n--) {
        if ($v === false) { return; }
        foreach ($v as $entry) {}
        $v = glob($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = glob($a);
        if ($v === false) { return; }
        foreach ($v as $entry) {}
    }
}
