<?php
function next_turn(string $a, string $b, $n) {
    $v = gzdecode($a);
    if ($v === false) { return; }
    while ($n--) {
        <warning descr="Check decompression before consuming data.">strlen($v)</warning>;
        $v = gzdecode($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = gzdecode($a);
    if ($v === false) { return; }
again:
    <warning descr="Check decompression before consuming data.">strlen($v)</warning>;
    $v = gzdecode($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = gzdecode($a);
    while ($n--) {
        if ($v === false) { return; }
        strlen($v);
        $v = gzdecode($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = gzdecode($a);
        if ($v === false) { return; }
        strlen($v);
    }
}
