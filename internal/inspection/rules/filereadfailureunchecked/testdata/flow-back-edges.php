<?php
function next_turn(string $a, string $b, $n) {
    $v = file_get_contents($a);
    if ($v === false) { return; }
    while ($n--) {
        <warning descr="Handle file reading failure before using its result.">trim($v)</warning>;
        $v = file_get_contents($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = file_get_contents($a);
    if ($v === false) { return; }
again:
    <warning descr="Handle file reading failure before using its result.">trim($v)</warning>;
    $v = file_get_contents($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = file_get_contents($a);
    while ($n--) {
        if ($v === false) { return; }
        trim($v);
        $v = file_get_contents($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = file_get_contents($a);
        if ($v === false) { return; }
        trim($v);
    }
}
