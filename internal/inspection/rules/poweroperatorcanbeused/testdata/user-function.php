<?php
namespace Geometry;

function pow($base, $exp) { return 0; }

$a = pow($side, 2);
$b = \Numbers\pow($side, 2);
$c = <warning descr="Use '$side ** 3' (exponentiation operator) instead.">\pow($side, 3)</warning>;
