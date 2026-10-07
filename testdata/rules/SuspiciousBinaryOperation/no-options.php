<?php
function f($p, $x, $y, $z) {
    $ok = [$p and FALSE, $p || true];
    if ($x || $y && $z) {}
}
