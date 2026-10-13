<?php
function order($o, $p, $q) {
    $a = load($o) && isset($p) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($q)</weak_warning> && $o->ready();
    $b = isset($p) && ($q = fetch()) && isset($q);
    $c = isset($p) && notify() && isset($q);
    $d = !isset($p) || $o->fail() || !isset($q);
    $e = isset($p) && f() && isset($q) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($o)</weak_warning>;
    $g = !isset($p) || $o === null || !<weak_warning descr="Merge this check into the preceding !isset() call.">isset($q)</weak_warning> || g();
    return [$a, $b, $c, $d, $e, $g];
}
