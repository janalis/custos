<?php
function recreated(): int {
    $im = imagecreate(5, 5);
    imagepalettetotruecolor($im);
    $im = imagecreate(5, 5);
    $c = imagecolorallocate($im, 10, 20, 30);
    return <warning descr="Look up palette components instead of shifting the index.">($c >> 16) & 255</warning>;
}

function otherImage(bool $flag): int {
    $im = imagecreate(5, 5);
    $other = imagecreate(5, 5);
    if ($flag) { imagepalettetotruecolor($other); }
    $c = imagecolorallocate($im, 10, 20, 30);
    return <warning descr="Look up palette components instead of shifting the index.">($c >> 8) & 255</warning>;
}

function convertedLater(bool $flag): int {
    $im = imagecreate(5, 5);
    $c = imagecolorallocate($im, 10, 20, 30);
    if ($flag) { imagepalettetotruecolor($im); }
    return <warning descr="Look up palette components instead of shifting the index.">($c >> 16) & 255</warning>;
}
