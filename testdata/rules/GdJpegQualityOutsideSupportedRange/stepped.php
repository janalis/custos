<?php
function later($i, $p) {
    $quality = 150;
    <error descr="Supply JPEG quality between zero and 100.">imagejpeg($i, $p, $quality)</error>;
    $quality--;
}
function other($i, $p) {
    $quality = 150;
    $count = 1;
    $count--;
    $copy = function () use ($quality) { $quality--; };
    $own = function () { $quality = 1; $quality--; };
    <error descr="Supply JPEG quality between zero and 100.">imagejpeg($i, $p, $quality)</error>;
}
