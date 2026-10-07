<?php
// level texts that are not plain decimal integers are kept as expressions
$big  = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 1 + 9999999999)'.">dirname(dirname($p), 9999999999)</warning>;
$dec  = <warning descr="Collapse the nested dirname() calls into 'dirname($p, 1 + (--$n))'.">dirname(dirname($p), --$n)</warning>;
