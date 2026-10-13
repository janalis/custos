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
    return !($a xor $b);
}
function both($a, $b) {
    return $a and $b;
}
