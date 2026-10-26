<?php
function size(string $png): int
{
    $image = \imagecreatefromstring($png);
    $w = \imagesx(<warning descr="Check image decoding success before using the image.">$image</warning>);
    $h = \imagesy($image);
    return $w * $h;
}

function guarded(string $png): int
{
    $image = imagecreatefromstring($png);
    try {
        $w = imagesx(<warning descr="Check image decoding success before using the image.">$image</warning>);
    } catch (\TypeError $e) {
        return 0;
    }
    return $w + imagesy(<warning descr="Check image decoding success before using the image.">$image</warning>);
}
