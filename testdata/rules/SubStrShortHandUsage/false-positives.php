<?php
function cut($name, $head, $k) {
    $f = substr($name, $k, strlen($name) - 1);
    $g = substr($name, 0, strlen($head) - 1);
    $h = substr($name, 0, (strlen($name) - 1));
    $i = substr($name, 0x1, strlen($name) - 2);
    $j = substr($name, 0, mb_strlen($name, 'UTF-8') - 1);
    $l = substr($name, 0, strlen($name) + 1);
    $m = substr($name, 0, 2 - strlen($name));
    $n = substr($name, 1);
    $o = substr($name, -3, strlen($name) - 1);
    $p = substr($name, 0, strlen($name) - - 2);
    // 'abcde': substr($s, 3, -1) is 'd', substr($s, 3, -3) is ''
    $q = substr($name, 3, strlen($name) - 6);
    $r = substr($name, 0, strlen($name) - 4);
    $q = substr($name, 1_0, strlen($name) - 2);
    return [$f, $g, $h, $i, $j, $l, $m, $n, $o, $p, $q, $s];
}
