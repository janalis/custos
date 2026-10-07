<?php
function pages(array $rows, $cursor, $iMax)
{
    for ($i = 0, $iMax1 = count($rows); $i < $iMax1; $i++) {
        echo $iMax;
    }
}

function nested(array $grid, $o)
{
    $loopsMax = 10;
    for ($o->r = 0, $loopsMax1 = count($grid); $o->r < $loopsMax1; $o->r++) {
        for ($o->c = 0, $loopsMax2 = count($grid[$o->r]); $o->c < $loopsMax2; $o->c++) {
            $o->cell($loopsMax);
        }
    }
}

function same(array $a)
{
    for ($k = 0, $kMax = count($a); $k < $kMax; $k++) {
        for ($k2 = 0, $kMax1 = strlen($a[$k]); $k < $kMax1; $k2++) {}
    }
    $f = function () use (&$jMax) {};
    for ($j = 0, $jMax1 = count($a); $j < $jMax1; $j++) {}
}
