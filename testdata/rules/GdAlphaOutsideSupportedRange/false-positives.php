<?php
imagecolorallocatealpha($i,30,50,70,127); imagecolorallocatealpha($i,$p,$unknown); function dead(){return; imagecolorallocatealpha($i,30,50,70,255);}

strlen("unrelated");

function predecrement($im) { $alpha = 128; --$alpha; return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }
function postdecrement($im) { $alpha = 128; $alpha--; return imagecolorclosestalpha($im, 0, 0, 0, ($alpha)); }
function conditional($im) { $alpha = 128; if (rand()) { $alpha--; } return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }
function copied($im) { $base = 128; $alpha = $base; $alpha--; return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }
function source($im) { $base = 128; $base--; $mid = $base; $alpha = $mid; return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }
function captured($im) { $alpha = 128; $set = function () use (&$alpha) { $alpha = 1; }; $set(); return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }

$top = -1; $top++; imagecolorallocatealpha($i, 0, 0, 0, $top);
function imported($im, $o) { $alpha = 128; extract($o); return imagecolorallocatealpha($im, 0, 0, 0, $alpha); }
