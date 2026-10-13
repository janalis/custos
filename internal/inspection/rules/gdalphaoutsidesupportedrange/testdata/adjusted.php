<?php
function later($im) { $alpha = 128; for ($i = 0; $i < 3; $i++) { imagecolorallocatealpha($im, 0, 0, 0, $alpha); $alpha--; } }
function arrow($im) { $alpha = 128; $step = fn() => --$alpha; return <error descr="Supply GD alpha between zero and 127.">imagecolorallocatealpha($im, 0, 0, 0, $alpha)</error>; }
function other($im) { $alpha = 200; $n = 0; $n++; $alpha = $alpha; $copy = function () use ($alpha, &$n) {}; return <error descr="Supply GD alpha between zero and 127.">imagecolorclosestalpha($im, 0, 0, 0, $alpha)</error>; }
function nested($im) { $alpha = -1; function () { $alpha = 1; $alpha++; }; return <error descr="Supply GD alpha between zero and 127.">imagecolorallocatealpha($im, 0, 0, 0, $alpha)</error>; }
function direct($im) { $alpha = 128; return <error descr="Supply GD alpha between zero and 127.">imagecolorallocatealpha($im, 0, 0, 0, alpha: 300)</error>; }
