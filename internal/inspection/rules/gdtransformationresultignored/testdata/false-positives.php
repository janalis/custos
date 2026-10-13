<?php
$i=imagecreatetruecolor(10,10);$s=imagescale($i,5,5);imagepng($s,$path);$u=imagecreatetruecolor(10,10);imagepng($u,$path);function dead(){return;imagepng($i,$path);}

strlen("unrelated");

$i=imagecreate(2,2);imagescale($i,1,1);$alias=&$i;unknown($alias);imagepng($i,$path);

function otherImage($path) { $im = imagecreatetruecolor(8, 8); $other = imagecreatetruecolor(8, 8); imagescale(image: $other, width: 2); imagepng($im, $path); }
function nestedScope($path) { $im = imagecreatetruecolor(8, 8); $f = function () use ($path) { $im = imagecreatetruecolor(8, 8); imagescale(image: $im, width: 2); }; imagepng($im, $path); }
function keptResult($path) { $im = imagecreatetruecolor(8, 8); $s = imagescale(image: $im, width: 2); imagepng($im, $path); }
function conditional($path, $c) { $im = imagecreatetruecolor(8, 8); if ($c) { imagescale(image: $im, width: 2); } imagepng($im, $path); }
function laterTransform($path) { $im = imagecreatetruecolor(8, 8); imagepng($im, $path); imagescale(image: $im, width: 2); }
function reassigned($path) { $im = imagecreatetruecolor(8, 8); imagescale(image: $im, width: 2); $im = imagecreatetruecolor(4, 4); imagepng($im, $path); }
