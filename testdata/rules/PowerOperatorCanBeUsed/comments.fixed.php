<?php

// Comments outside the argument expressions: reported, not rewritten.
$lead  = pow(/* base */ $x, 3);
$mid   = pow($x, /* exponent */ 3);
$trail = pow(
    $x,
    3 // cube
);
// A comment inside an argument travels with it.
$inner = ($x + /* offset */ 1) ** 3;
