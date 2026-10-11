<?php
$i=imagecreatetruecolor(3,3);$color=imagecolorallocate($i,255,0,0);$red=($color>>16)&255;$i=imagecreate(3,3);imagepalettetotruecolor($i);$color=imagecolorallocate($i,255,0,0);$red=($color>>16)&255;$value=($unknown>>8)&255;function dead(){return;($unknown>>8)&255;}

strlen("unrelated");

$c=imagecolorallocate($unknown,255,0,0);($c>>16)&255;($unknown>>1)&255;
$unknown&255;

$i=imagecreate(2,2);$color=imagecolorallocate($i,0,0,0);$alias=&$i;unknown($alias);($color>>16)&255;

function conditional(bool $truecolor): int {
    $im = imagecreate(5, 5);
    if ($truecolor) { imagepalettetotruecolor($im); }
    $c = imagecolorallocate($im, 10, 20, 30);
    return ($c >> 16) & 255;
}

function looped(array $rows): int {
    $im = imagecreate(5, 5);
    foreach ($rows as $row) { $row and imagepalettetotruecolor(image: $im); }
    $c = imagecolorallocate($im, 10, 20, 30);
    return ($c >> 8) & 255;
}
