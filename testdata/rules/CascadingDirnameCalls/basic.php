<?php
$base = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 3)'.">dirname(dirname(dirname($root)))</warning>;
$app  = <warning descr="Collapse the nested dirname() calls into 'dirname(getcwd(), 4)'.">dirname(dirname(dirname(dirname(getcwd()))))</warning>;
$var  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 1 + $up)'.">dirname(dirname($root, $up))</warning>;
$sum  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 5)'.">dirname(dirname($root, 2), 3)</warning>;
$two  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 0 + $b + $a)'.">dirname(dirname($root, $a), $b)</warning>;
$fq   = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 2)'.">\dirname(\dirname($root))</warning>;
$tern = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 1 + ($deep ? 2 : 1))'.">dirname(dirname($root, $deep ? 2 : 1))</warning>;
$oct  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 1 + 010)'.">dirname(dirname($root, 010))</warning>;
$stop = <warning descr="Collapse the nested dirname() calls into 'dirname($fs->dirname($p), 2)'.">dirname(dirname($fs->dirname($p)))</warning>;
