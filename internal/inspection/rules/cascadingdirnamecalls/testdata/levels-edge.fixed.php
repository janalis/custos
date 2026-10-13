<?php
// level texts that are not plain decimal integers are kept as expressions
$big  = dirname($p, 1 + 9999999999);
$dec  = dirname($p, 1 + (--$n));
// levels keep the evaluation order of the nested form: inner first
$ord  = dirname($p, 0 + ($i = 1) + ($i + 1));
$call = dirname($p, 1 + inner() + outer());
