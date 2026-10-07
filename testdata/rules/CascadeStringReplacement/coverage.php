<?php
define('SEP', '-');
define('PAIR', ['+', '*']);
define('TWICE', 'a');
define('TWICE', 'b');
define('ODD');
define($dynamic, 'x');

function edges($s, $r) {
    if ($s === '') {
        return;
    }
    $n = str_replace('a', 'b');
    log_it($s = str_replace('c', 'd', $s));
    log_it($s);
    $s = str_replace('e', 'f', $s);
    return $s;
}

function afterReturn($s) {
    return $s;
    $s = str_replace('a', 'b', $s);
}

function keyed($s, $t, $u, $v, $w, $x, $y, $z, array $more, &$ref) {
    $s = str_replace('p', 'q', $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([...$more, 'k' => 'x'], 'y', $s)</warning>;

    $t = str_replace('p', 'q', $t);
    $t = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['k' => 'a', 'b'], 'y', $t)</warning>;

    $u = str_replace('p', 'q', $u);
    $u = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['1.5' => 'a', 1.5 => 'b'], 'y', $u)</warning>;

    $v = str_replace('p', 'q', $v);
    $v = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['-1' => 'a', 'b'], 'y', $v)</warning>;

    $w = str_replace('p', 'q', $w);
    $w = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([<<<K
    key
    K => 'a', 'b'], 'y', $w)</warning>;

    $x = str_replace('p', 'q', $x);
    $x = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['k' => &$ref, 2 => 'b', 'c'], 'y', $x)</warning>;

    $y = str_replace('p', 'q', $y);
    $y = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([], [], $y)</warning>;

    $z = str_replace('p', [$r = 'q'], $z);
    $z = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('o', [$r], $z)</warning>;
    return $s . $t . $u . $v . $w . $x . $y . $z;
}

function constants($s, $t, $u, $v) {
    $s = str_replace('a', SEP, $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('b', SEP, $s)</warning>;

    $t = str_replace('a', TWICE, $t);
    $t = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('b', TWICE, $t)</warning>;

    $u = str_replace('a', true, $u);
    $u = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('b', true, $u)</warning>;

    $v = str_replace(['a', 'b'], PAIR, $v);
    $v = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c', 'd'], ODD, $v)</warning>;
    return $s . $t . $u . $v;
}

function longArrays($s) {
    $s = str_replace('a', 'b', $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(array( 'c', 'd' ), array( 'e', 'f' ), $s)</warning>;
    return $s;
}

function keyedSearch($s, $t, &$r) {
    $a = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">['x' => 'a', 'y' => 'a']</weak_warning>, 'b', $s);
    $b = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">['-1' => 'a']</weak_warning>, 'b', $t);
    $c = str_replace(['a', 'a' => 'b'], 'c', $s);
    $d = str_replace([&$r, &$r], 'c', $s);
    return [$a, $b, $c, $d];
}
