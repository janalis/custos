<?php
function elseGuard($a, $x) {
    if ($x) { return 1; } else { $a++; }
    <warning descr="Return the condition directly: 'return $a > 1'.">if</warning> ($a > 1) { return true; }
    return false;
}
function emptyGuard($a, $x) {
    if ($x) {}
    <warning descr="Return the condition directly: 'return $a > 1'.">if</warning> ($a > 1) { return true; }
    return false;
}
