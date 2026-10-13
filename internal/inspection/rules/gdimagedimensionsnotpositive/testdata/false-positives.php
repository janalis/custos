<?php
imagecreate(3,3);imagecreatetruecolor($x,$y);function dead(){return;imagecreate(0,0);}

strlen("unrelated");

function countedRows(array $lines) { $rows = 0; foreach ($lines as $line) { $rows++; } return imagecreatetruecolor(200, $rows); }
function preIncremented() { $w = 0; ++$w; return imagecreate($w, 5); }
function copiedCounter() { $n = 0; $n++; $m = $n; return imagecreate($m, 5); }
function capturedByReference() { $n = 0; $f = function () use (&$n) { $n = 10; }; $f(); return imagecreatetruecolor(200, $n); }
function caught() { $n = 0; try { load(); } catch (\Exception $n) { } return imagecreate(5, $n); }
function extracted(array $data) { $n = 0; extract($data); return imagecreate($n, 5); }
function dynamicCounter($name) { $n = 0; $$name++; return imagecreate($n, 5); }
function dynamicCompound($name) { $n = 0; $$name += 4; return imagecreate($n, 5); }
function sharedGlobal() { global $n; $n = 0; resize(); return imagecreate($n, 5); }
function sharedStatic() { static $n; $n = 0; $g = function () {}; return imagecreate($n, 5); }
function sharedParameter(&$n) { $n = 0; Sizer::grow(); return imagecreate($n, 5); }
$shared = 0; $outer = 0;
$resize = function () use (&$outer) { $outer = 0; new Sizer(); return imagecreate($outer, 5); };
$sizer->grow();
imagecreate($shared, 5);
