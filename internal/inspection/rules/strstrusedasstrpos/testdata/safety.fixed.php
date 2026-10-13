<?php
function probe($h, $n, $flag) {
    $r = [];
    $r[] = (strpos($h, $n) === false) == $flag;
    $r[] = (int) (strpos($h, $n) === false);
    $r[] = !(strpos($h, $n) === false);
    $r[] = (stripos($h, $n) === false) . 'x';
    $r[] = (strpos($h, $n) === false) + 1;
    $r[] = -(strpos($h, $n) === false);
    $r[] = @(strpos($h, $n) === false);
    $r[] = (strpos($h, $n) === false) <=> 1;
    $r[] = clone (strpos($h, $n) === false);
    $r[] = strpos($h, $n) === false ?: 'z';
    $r[] = strpos($h, $n) === false ?? 'z';
    $r[] = $flag & strpos($h, $n) === false;
    $r[] = [strpos($h, $n) === false];
    $r[] = strpos($h, $n) !== false ? 1 : 2;
    $r[] = strpos($h, $n) === false;
    $r[] = stripos($h, $n) !== false;
    $r[] = strpos($h, $n) === false;
    return $r;
}
