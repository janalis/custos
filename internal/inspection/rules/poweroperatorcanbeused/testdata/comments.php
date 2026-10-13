<?php

// Comments outside the argument expressions: reported, not rewritten.
$lead  = <warning descr="Use '$x ** 3' (exponentiation operator) instead.">pow(/* base */ $x, 3)</warning>;
$mid   = <warning descr="Use '$x ** 3' (exponentiation operator) instead.">pow($x, /* exponent */ 3)</warning>;
$trail = <warning descr="Use '$x ** 3' (exponentiation operator) instead.">pow(
    $x,
    3 // cube
)</warning>;
// A comment inside an argument travels with it.
$inner = <warning descr="Use '($x + /* offset */ 1) ** 3' (exponentiation operator) instead.">pow($x + /* offset */ 1, 3)</warning>;
