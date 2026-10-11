<?php
imagejpeg($i,$p,100); imagejpeg($i,$p,$unknown); function dead(){return; imagejpeg($i,$p,150);}

strlen("unrelated");

function shrink($i, $p) {
    $quality = 101;
    do {
        $quality--;
        imagejpeg($i, $p, $quality);
    } while ($quality > 10);
}
function raise($i, $p) {
    $quality = -2;
    ++$quality;
    imagejpeg($i, $p, $quality);
}
function shared($i, $p) {
    $quality = 101;
    $lower = function () use (&$quality) { $quality--; };
    $lower();
    imagejpeg($i, $p, $quality);
}
$top = 101;
$top--;
imagejpeg($i, $p, $top);
function copied($i, $p) {
    $base = 101;
    $base--;
    $quality = $base;
    imagejpeg($i, $p, $quality);
}
function reassigned($i, $p) {
    $quality = 101;
    $lower = function () use (&$quality) { $quality = 50; };
    $lower();
    imagejpeg($i, $p, $quality);
}
function imported($i, $p, $o, $s) {
    $quality = 101;
    extract($o);
    imagejpeg($i, $p, $quality);
}
function parsed($i, $p, $s) {
    $quality = 101;
    parse_str($s);
    imagejpeg($i, $p, $quality);
}
