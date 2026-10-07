<?php
$one  = dirname($root);
$lvl  = dirname($root, 1);
$far  = dirname($root, 4);
$wrap = dirname(getcwd());
$real = dirname(realpath($root), 2);
$zero = dirname(dirname($root, 0));
$bad  = dirname(dirname());
$none = dirname();
$many = dirname(dirname($root, 1, 2));
$par  = dirname((dirname($root)));
