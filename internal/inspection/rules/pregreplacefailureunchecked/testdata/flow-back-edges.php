<?php
function next_turn(string $a, string $b, $n) {
    $v = preg_replace('/x/', 'y', $a);
    if ($v === null) { return; }
    while ($n--) {
        <warning descr="Handle regex replacement failure before using its result.">trim($v)</warning>;
        $v = preg_replace('/x/', 'y', $b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = preg_replace('/x/', 'y', $a);
    if ($v === null) { return; }
again:
    <warning descr="Handle regex replacement failure before using its result.">trim($v)</warning>;
    $v = preg_replace('/x/', 'y', $b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = preg_replace('/x/', 'y', $a);
    while ($n--) {
        if ($v === null) { return; }
        trim($v);
        $v = preg_replace('/x/', 'y', $b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = preg_replace('/x/', 'y', $a);
        if ($v === null) { return; }
        trim($v);
    }
}
