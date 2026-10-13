<?php
function bounds(array $retry, array $rows)
{
    // One iteration more than the elements: not a foreach.
    for ($a = 0; $a <= count($retry); $a++) {
        echo $retry[$a] ?? 0;
    }
    for ($b = 0; count($rows) >= $b; $b++) {
        echo $rows[$b];
    }
    for ($c = 0; $c + count($rows); $c++) {
        echo $rows[$c];
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($d = 0; $d != count($rows); $d++) {
        echo $rows[$d];
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($e = 0; count($rows) !== $e; $e++) {
        echo $rows[$e];
    }
}
