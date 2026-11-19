<?php

class Point {}

function arrays(array $row, int $n, $unknown)
{
    $a = <error descr="Use a number here, not array.">$row</error> - $n;
    $b = <error descr="Use a number here, not array.">$row</error> * $unknown;
    $c = $n / <error descr="Use a number here, not array.">$row</error>;
    $d = <error descr="Use a number here, not array.">$row</error> % 2;
    $e = <error descr="Use a number here, not array.">$row</error> ** 2;
    $f = <error descr="Use a number here, not array.">$row</error> & 1;
    $g = <error descr="Use a number here, not array.">$row</error> | 1;
    $h = <error descr="Use a number here, not array.">$row</error> ^ 1;
    $i = <error descr="Use a number here, not array.">$row</error> << 1;
    $j = <error descr="Use a number here, not array.">$row</error> >> 1;
    $k = <error descr="Use a number here, not array.">$row</error> + $n;
    $l = $n + <error descr="Use a number here, not array.">$row</error>;
}

function objects(Point $point, int $n)
{
    $a = <error descr="Use a number here, not Point.">$point</error> + $n;
    $b = $n * <error descr="Use a number here, not Point.">$point</error>;
}

function compound(array $a, array $b, array $c, array $d, array $e, array $f, array $g, array $h, array $i, array $j, array $k, Point $point, int $n)
{
    <error descr="Use a number here, not array.">$a</error> += 1;
    <error descr="Use a number here, not array.">$b</error> -= 1;
    <error descr="Use a number here, not array.">$c</error> *= 1;
    <error descr="Use a number here, not array.">$d</error> /= 1;
    <error descr="Use a number here, not array.">$e</error> %= 1;
    <error descr="Use a number here, not array.">$f</error> **= 1;
    <error descr="Use a number here, not array.">$g</error> &= 1;
    <error descr="Use a number here, not array.">$h</error> |= 1;
    <error descr="Use a number here, not array.">$i</error> ^= 1;
    <error descr="Use a number here, not array.">$j</error> <<= 1;
    <error descr="Use a number here, not array.">$k</error> >>= 1;
    $n += <error descr="Use a number here, not Point.">$point</error>;
}
