<?php
function conditions(array $a, array $b, $ok)
{
    $acc = [];
    $acc = array_merge($acc, $a);
    for ($i = 0; $ok, <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>; $i++) {}
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($a)</error>, <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i < count($b)</error>; $i++) {}
    return $acc;
}

function once(array $a)
{
    $acc = [];
    $acc = array_merge($acc, $a);
    log_it($acc);
}
