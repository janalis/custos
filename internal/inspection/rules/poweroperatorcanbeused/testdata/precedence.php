<?php

function squares()
{
    $sent = <warning descr="Use '(yield 3) ** 2' (exponentiation operator) instead.">pow(yield 3, 2)</warning>;
    $from = <warning descr="Use '2 ** (yield from inner())' (exponentiation operator) instead.">pow(2, yield from inner())</warning>;
    return $sent + $from;
}

$shown  = <warning descr="Use '(print 3) ** 2' (exponentiation operator) instead.">pow(print 3, 2)</warning>;
$loaded = <warning descr="Use '(include $path) ** 2' (exponentiation operator) instead.">pow(include $path, 2)</warning>;
$lazy   = <warning descr="Use '$n ** (fn() => 2)' (exponentiation operator) instead.">pow($n, fn() => 2)</warning>;
$digit  = <warning descr="Use '(2 ** 3)' (exponentiation operator) instead.">pow(2, 3)</warning>[0] ?? 'none';
$prop   = <warning descr="Use '($big ** 2)' (exponentiation operator) instead.">pow($big, 2)</warning>->num ?? 'none';
$isNum  = <warning descr="Use '($n ** !$flag)' (exponentiation operator) instead.">pow($n, !$flag)</warning> instanceof GMP;
$copy   = clone <warning descr="Use '($big ** 2)' (exponentiation operator) instead.">pow($big, 2)</warning>;
$index  = $table[<warning descr="Use '$n ** 2' (exponentiation operator) instead.">pow($n, 2)</warning>];
$signed = -<warning descr="Use '$n ** 2' (exponentiation operator) instead.">pow($n, 2)</warning>;
