<?php
function next_turn(string $a, string $b, $n) {
    $v = parse_url($a);
    if ($v === false) { return; }
    while ($n--) {
        echo <warning descr="Reject URL parsing failure before indexing.">$v['host']</warning>;
        $v = parse_url($b);
    }
}
function backward_jump(string $a, string $b, $n) {
    $v = parse_url($a);
    if ($v === false) { return; }
again:
    echo <warning descr="Reject URL parsing failure before indexing.">$v['host']</warning>;
    $v = parse_url($b);
    if ($n--) { goto again; }
}
function checked_each_turn(string $a, string $b, $n) {
    $v = parse_url($a);
    while ($n--) {
        if ($v === false) { return; }
        echo $v['host'];
        $v = parse_url($b);
    }
}
function assigned_before_every_use(string $a, $n) {
    while ($n--) {
        $v = parse_url($a);
        if ($v === false) { return; }
        echo $v['host'];
    }
}
