<?php
function sum($a, $b) {
    if ($a + $b) { return true; }
    return false;
}
function joined($a, $b) {
    if ($a . $b) { return true; } else { return false; }
}
function fallback($a, $b) {
    if (($a ?? $b)) { return false; }
    return true;
}
function masked($flags) {
    if ($flags & 4) { return true; }
    return false;
}
function ordered($a, $b) {
    if ($a <=> $b) { return true; }
    return false;
}
function either($a, $b) {
    <warning descr="Return the condition directly: 'return !($a xor $b)'.">if</warning> ($a xor $b) { return false; }
    return true;
}
function both($a, $b) {
    <warning descr="Return the condition directly: 'return $a and $b'.">if</warning> ($a and $b) { return true; }
    return false;
}
