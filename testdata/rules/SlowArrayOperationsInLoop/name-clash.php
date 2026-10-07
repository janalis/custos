<?php
function pages(array $rows, $cursor, $iMax)
{
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($rows)</error>; $i++) {
        echo $iMax;
    }
}

function nested(array $grid, $o)
{
    $loopsMax = 10;
    for ($o->r = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$o->r < count($grid)</error>; $o->r++) {
        for ($o->c = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$o->c < count($grid[$o->r])</error>; $o->c++) {
            $o->cell($loopsMax);
        }
    }
}

function same(array $a)
{
    for ($k = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$k < count($a)</error>; $k++) {
        for ($k2 = 0; <error descr="'strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">$k < strlen($a[$k])</error>; $k2++) {}
    }
    $f = function () use (&$jMax) {};
    for ($j = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$j < count($a)</error>; $j++) {}
}
