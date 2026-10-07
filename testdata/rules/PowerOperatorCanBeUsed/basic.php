<?php
$area   = <warning descr="Use '$side ** 2' (exponentiation operator) instead.">pow($side, 2)</warning>;
$growth = <warning descr="Use '($rate * 0.01 + 1) ** $years' (exponentiation operator) instead.">pow($rate * 0.01 + 1, $years)</warning>;
$scaled = <warning descr="Use '2 ** ($bits - 1)' (exponentiation operator) instead.">\pow(2, $bits - 1)</warning>;
$total  = $offset - <warning descr="Use '($k ** $m)' (exponentiation operator) instead.">pow($k, $m)</warning>;
$safe   = <warning descr="Use '($q ?? 0) ** ($flag ? 2 : 3)' (exponentiation operator) instead.">pow($q ?? 0, $flag ? 2 : 3)</warning>;
$short  = <warning descr="Use '$x ** ($n ?: 2)' (exponentiation operator) instead.">pow($x, $n ?: 2)</warning>;
$neg    = <warning descr="Use '(-$x) ** 2' (exponentiation operator) instead.">pow(-$x, 2)</warning>;
$inc    = <warning descr="Use '(++$i) ** 2' (exponentiation operator) instead.">pow(++$i, 2)</warning>;
$post   = <warning descr="Use '$i++ ** --$j' (exponentiation operator) instead.">pow($i++, --$j)</warning>;
$other  = max($side, 2);
$dyn    = $fn($side, 2);
$keep   = pow($side);
$spread = pow(...$pair);
$meth   = $calc->pow($side, 2);
$stat   = Calc::pow($side, 2);
