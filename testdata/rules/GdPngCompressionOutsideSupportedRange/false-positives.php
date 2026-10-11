<?php
imagepng($i,$p,9); imagepng($i,$p,$unknown); function dead(){return; imagepng($i,$p,90);}

strlen("unrelated");

function decremented($i, $p) { $level = 10; $level--; imagepng($i, $p, $level); }
function captured($i, $p) { $level = 10; $g = function () use (&$level) { $level = 3; }; $g(); imagepng($i, $p, $level); }
function referenced($i, $p) { $level = 10; foreach ([&$level] as &$q) { $q = 3; } imagepng($i, $p, $level); }
function adjusted($i, $p, $o) { $level = 10; $o->clampLevel($level); imagepng($i, $p, $level); }
function imported($i, $p, $o) { $level = 10; extract($o); imagepng($i, $p, $level); }
function copied($i, $p) { $base = 10; $base--; $level = $base; imagepng($i, $p, $level); }
