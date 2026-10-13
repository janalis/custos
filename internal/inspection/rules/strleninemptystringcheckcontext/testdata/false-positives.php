<?php
function probe(string $s, $obj)
{
    $a = 1 > strlen($s);
    $b = 1 <= strlen($s);
    $c = strlen($s) == 1;
    $d = strlen($s) >= 0;
    $e = strlen($s) === 0.0;
    $f = strlen($s) == '0';
    $g = strlen($s) > LIMIT;
    $h = strlen($s) + 1;
    for (; strlen($s); ) {}
    $i = strlen($s) xor true;
    $j = $a ? strlen($s) : 0;
    $k = (bool) strlen($s);
    $l = $obj->strlen($s);
    $m = Str::strlen($s);
    $n = strlen();
    $o = strlen($s) > -0;
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l, $m, $n, $o];
}

