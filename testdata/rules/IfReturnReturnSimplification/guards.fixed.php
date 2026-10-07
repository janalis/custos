<?php
function elseGuard($a, $x) {
    if ($x) { return 1; } else { $a++; }
    return $a > 1;
}
function emptyGuard($a, $x) {
    if ($x) {}
    return $a > 1;
}
