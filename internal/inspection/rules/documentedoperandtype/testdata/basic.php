<?php

class Point {}

/**
 * @param int[] $counts
 * @param Point $point
 * @param array<string, int> $row
 */
function documented($counts, $point, $row, int $n)
{
    $a = <warning descr="Use a number here, not int[].">$counts</warning> * 2;
    $b = $n - <warning descr="Use a number here, not Point.">$point</warning>;
    $c = <warning descr="Use a number here, not int[].">$row</warning> + $n;
    <warning descr="Use a number here, not int[].">$counts</warning> += 1;
}
