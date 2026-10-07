<?php
$a = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 2)'.">Dirname(dirname($root, 1))</warning>;
$b = <warning descr="Collapse the nested dirname() calls into 'dirname(__FILE__, 3)'.">DIRNAME(DirName(dirname(__FILE__)))</warning>;
$c = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 2)'.">\DIRNAME(\dirname($root))</warning>;
$d = DIRNAME($root);
