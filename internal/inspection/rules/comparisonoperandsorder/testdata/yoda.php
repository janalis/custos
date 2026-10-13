<?php

$isZero  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$counter === 0</weak_warning>;
$isNeg   = <weak_warning descr="Put the constant operand on the left side of the comparison.">$delta != -3</weak_warning>;
$isOn    = <weak_warning descr="Put the constant operand on the left side of the comparison.">$flag == true</weak_warning>;
$isUnix  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$sep !== "/"</weak_warning>;
$fileOk  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$path <> __FILE__</weak_warning>;
$interp  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$name   ==   "x{$y}"</weak_warning>;
$ns      = <weak_warning descr="Put the constant operand on the left side of the comparison.">$v === \App\LIMIT</weak_warning>;

// no report
$same    = 'a' === $label;
$both    = PHP_INT_MAX == 99;
$classes = $mode === Mode::FAST;
$less    = $counter < 5;
$paren   = $x == (5);
$plus    = $x == +5;
$arr     = $x == [];
