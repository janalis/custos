<?php
function counters(array $xs)
{
    $n = 0;
    foreach ($xs as $x) {
        ++$n;
    }
    echo $n[0];

    $total = 1.5;
    $total *= 2;
    return $total['sum'];
}
