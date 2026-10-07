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
    $s = str_replace([...$more, 'k' => 'x'], 'y', $s);

    $t = str_replace(array('p', 'a', 'b'), array('q', 'y', 'y'), $t);

    $u = str_replace('p', 'q', $u);
    $u = str_replace(['1.5' => 'a', 1.5 => 'b'], 'y', $u);

    $v = str_replace('p', 'q', $v);
    $v = str_replace(['-1' => 'a', 'b'], 'y', $v);

    $w = str_replace('p', 'q', $w);
    $w = str_replace([<<<K
    key
    K => 'a', 'b'], 'y', $w);

    $x = str_replace(array('p', &$ref, 'b', 'c'), array('q', 'y', 'y', 'y'), $x);

    $y = str_replace(array('p'), array('q'), $y);

    $z = str_replace(array('p', 'o'), array($r = 'q', $r), $z);
    return $s . $t . $u . $v . $w . $x . $y . $z;
}

function constants($s, $t, $u, $v) {
    $s = str_replace(array('a', 'b'), SEP, $s);

    $t = str_replace(array('a', 'b'), array(TWICE, TWICE), $t);

    $u = str_replace(array('a', 'b'), array(true, true), $u);

    $v = str_replace(array('a', 'b', 'c', 'd'), array(...array_values(PAIR), ODD, ODD), $v);
    return $s . $t . $u . $v;
}

function longArrays($s) {
    $s = str_replace(array( 'a', 'c', 'd' ), array( 'b', 'e', 'f' ), $s);
    return $s;
}

function keyedSearch($s, $t, &$r) {
    $a = str_replace('a', 'b', $s);
    $b = str_replace('a', 'b', $t);
    $c = str_replace(['a', 'a' => 'b'], 'c', $s);
    $d = str_replace([&$r, &$r], 'c', $s);
    return [$a, $b, $c, $d];
}
