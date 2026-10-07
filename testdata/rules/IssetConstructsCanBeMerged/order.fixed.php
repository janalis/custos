<?php
function order($o, $p, $q) {
    $a = load($o) && isset($p, $q) && $o->ready();
    $b = isset($p) && ($q = fetch()) && isset($q);
    $c = isset($p) && notify() && isset($q);
    $d = !isset($p) || $o->fail() || !isset($q);
    $e = isset($p) && f() && isset($q, $o);
    $g = !isset($p, $q) || $o === null || g();
    return [$a, $b, $c, $d, $e, $g];
}
