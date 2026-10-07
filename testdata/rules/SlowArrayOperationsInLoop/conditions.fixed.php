<?php
function conditions(array $a, array $b, $ok)
{
    $acc = [];
    $acc = array_merge($acc, $a);
    for ($i = 0, $iMax = count($a); $ok, $i < $iMax; $i++) {}
    for ($i = 0, $iMax = count($a); $i < $iMax, $i < count($b); $i++) {}
    return $acc;
}

function once(array $a)
{
    $acc = [];
    $acc = array_merge($acc, $a);
    log_it($acc);
}
