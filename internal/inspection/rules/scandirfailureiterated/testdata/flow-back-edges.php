<?php
function next_turn(string $a, string $b, $n) {
    $v = scandir($a);
    if ($v === false) { return; }
    while ($n--) {
        foreach (<warning descr="Check directory enumeration before iterating.">$v</warning> as $entry) {}
        $v = scandir($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = scandir($a);
    if ($v === false) { return; }
again:
    foreach (<warning descr="Check directory enumeration before iterating.">$v</warning> as $entry) {}
    $v = scandir($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = scandir($a);
    while ($n--) {
        if ($v === false) { return; }
        foreach ($v as $entry) {}
        $v = scandir($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = scandir($a);
        if ($v === false) { return; }
        foreach ($v as $entry) {}
    }
}
