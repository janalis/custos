<?php
// level texts that are not plain decimal integers are kept as expressions
$big  = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 1 + 9999999999)'.">dirname(dirname($p), 9999999999)</warning>;
$dec  = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 1 + (--$n))'.">dirname(dirname($p), --$n)</warning>;
// levels keep the evaluation order of the nested form: inner first
$ord  = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 0 + ($i = 1) + ($i + 1))'.">dirname(dirname($p, $i = 1), $i + 1)</warning>;
$call = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 1 + inner() + outer())'.">dirname(dirname(dirname($p, inner())), outer())</warning>;
