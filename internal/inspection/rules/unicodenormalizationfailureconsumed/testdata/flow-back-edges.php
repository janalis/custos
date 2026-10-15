<?php
function next_turn(string $a, string $b, $n) {
    $v = normalizer_normalize($a);
    if ($v === false) { return; }
    while ($n--) {
        <warning descr="Check normalization before consuming text.">strlen($v)</warning>;
        $v = normalizer_normalize($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = normalizer_normalize($a);
    if ($v === false) { return; }
again:
    <warning descr="Check normalization before consuming text.">strlen($v)</warning>;
    $v = normalizer_normalize($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = normalizer_normalize($a);
    while ($n--) {
        if ($v === false) { return; }
        strlen($v);
        $v = normalizer_normalize($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = normalizer_normalize($a);
        if ($v === false) { return; }
        strlen($v);
    }
}
