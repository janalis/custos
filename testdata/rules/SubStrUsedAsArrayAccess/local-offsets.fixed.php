<?php
// A local offset nothing else can reach keeps its assigned value.
function firstByte(string $s, array $rows) {
    $at = 0;
    $copy = function () use ($at) { return $at - 2; };
    $pairs = ['k' => $at, &$rows];
    $copy();
    return ($s[$at] ?? '');
}

function lastByte(string $s, int &$other) {
    static $calls;
    global $registry;
    $end = -1;
    $other++;
    return ($s[strlen($s) - 1] ?? '');
}
