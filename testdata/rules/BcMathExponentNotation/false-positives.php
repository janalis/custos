<?php
$a = bcpow('2', '100');
$b = bcpowmod('2', '10', '7');
// The scale is an integer, not a numeric string operand.
$c = bcsqrt('4', 2);
$d = bcpow('2', $exponent);
