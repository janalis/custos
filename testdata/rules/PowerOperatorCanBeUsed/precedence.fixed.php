<?php

function squares()
{
    $sent = (yield 3) ** 2;
    $from = 2 ** (yield from inner());
    return $sent + $from;
}

$shown  = (print 3) ** 2;
$loaded = (include $path) ** 2;
$lazy   = $n ** (fn() => 2);
$digit  = (2 ** 3)[0] ?? 'none';
$prop   = ($big ** 2)->num ?? 'none';
$isNum  = ($n ** !$flag) instanceof GMP;
$copy   = clone ($big ** 2);
$index  = $table[$n ** 2];
$signed = -$n ** 2;
