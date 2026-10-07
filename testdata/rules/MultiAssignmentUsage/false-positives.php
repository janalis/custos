<?php
function g(array $rows, $v) {
    list($a, $b) = $v;
    foreach ($rows as $row) {
        while (list($a) = $row) {}
        use_it(list($a) = $row);
        list($a) = $row[0];
        list($a) = $row->x;
        $fn = function () use ($row) { list($u) = $row; };
        $fn2 = fn() => [$q] = $row;
    }
    $x = $rows[0];
    $y = $other[1];
    $s = $rows['a'];
    $t = $rows['b'];
    $k = $rows[$i];
    $l = $rows[$j];
    $m = $rows[ONE];
    $n = $rows[TWO];
    $o = $rows[0];
    $o .= $rows[1];
    $out[1] = $rows[0];
    $out[2] = $rows[1];
    $w = $other[0];
    echo 1;
    $z = $rows[1];
}
function sameKey(array $r) {
    $p = $r[0];
    $q = $r[0];
    echo 1;
    $s = $r[1];
    $t = $r[1.0];
    echo 1;
    $u = $r[0x2];
    $v = $r[2];
}
