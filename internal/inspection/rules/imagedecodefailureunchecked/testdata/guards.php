<?php

function wrongArm(string $bytes): int
{
    $im = imagecreatefromstring($bytes);
    return $im ? 0 : imagesx(<warning descr="Check image decoding success before using the image.">$im</warning>);
}

function staleGuard(string $bytes, string $other): void
{
    $im = imagecreatefromstring($bytes);
    if (!$im) {
        return;
    }
    $im = imagecreatefromstring($other);
    imagepng(<warning descr="Check image decoding success before using the image.">$im</warning>);
}

function logged(string $bytes): void
{
    $im = imagecreatefromstring($bytes);
    if ($im === false) {
        error_log('Invalid image');
    }
    imagepng(<warning descr="Check image decoding success before using the image.">$im</warning>);
}
