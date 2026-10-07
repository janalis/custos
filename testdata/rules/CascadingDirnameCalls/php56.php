<?php
$base = dirname(dirname(dirname($root)));
$app  = dirname(dirname(dirname(dirname(getcwd()))));
$var  = dirname(dirname($root, $up));
$sum  = dirname(dirname($root, 2), 3);
$two  = dirname(dirname($root, $a), $b);
$fq   = \dirname(\dirname($root));
$tern = dirname(dirname($root, $deep ? 2 : 1));
$oct  = dirname(dirname($root, 010));
$stop = dirname(dirname($fs->dirname($p)));
