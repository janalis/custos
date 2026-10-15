<?php
function next_turn(string $a, string $b, $n) {
    $v = fopen($a, 'r');
    if ($v === false) { return; }
    while ($n--) {
        <warning descr="Check that the stream opened successfully.">fread($v, 12)</warning>;
        $v = fopen($b, 'r');
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = fopen($a, 'r');
    if ($v === false) { return; }
again:
    <warning descr="Check that the stream opened successfully.">fread($v, 12)</warning>;
    $v = fopen($b, 'r');
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = fopen($a, 'r');
    while ($n--) {
        if ($v === false) { return; }
        fread($v, 12);
        $v = fopen($b, 'r');
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = fopen($a, 'r');
        if ($v === false) { return; }
        fread($v, 12);
    }
}
