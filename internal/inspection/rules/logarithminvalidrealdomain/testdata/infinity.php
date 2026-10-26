<?php
$negInf = log(0);
$posInf = -log(0.0);
$tenth = log10((0));
$edge = log1p(-1);
$edgeFloat = log1p(-1.0);
$zero = 0;
$fromVar = <warning descr="Use a positive logarithm argument.">log($zero)</warning>;
$below = <warning descr="Use a positive logarithm argument.">log1p(-2)</warning>;
$neg = <warning descr="Use a positive logarithm argument.">log(-0.5)</warning>;
$base = <warning descr="Use a positive logarithm argument.">log(0, 1)</warning>;
$hex = <warning descr="Use a positive logarithm argument.">log(0x0)</warning>;
