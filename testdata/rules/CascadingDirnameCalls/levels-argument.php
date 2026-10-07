<?php
$up   = dirname($root, <warning descr="Collapse the nested dirname() calls into 'dirname($depth, 2)'.">dirname(dirname($depth))</warning>);
$keep = dirname($root, dirname($depth, 2));
