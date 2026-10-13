<?php
// A local offset nothing else can reach keeps its assigned value.
function firstByte(string $s, array $rows) {
    $at = 0;
    $copy = function () use ($at) { return $at - 2; };
    $pairs = ['k' => $at, &$rows];
    $copy();
    return <warning descr="Use '($s[$at] ?? '')' (string offset access) instead.">substr($s, $at, 1)</warning>;
}

function lastByte(string $s, int &$other) {
    static $calls;
    global $registry;
    $end = -1;
    $other++;
    return <warning descr="Use '($s[strlen($s) - 1] ?? '')' (string offset access) instead.">substr($s, $end, 1)</warning>;
}
