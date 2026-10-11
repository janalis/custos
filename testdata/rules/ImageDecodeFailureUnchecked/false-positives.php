<?php

function thumbnail(string $bytes): void
{
    $im = imagecreatefromstring($bytes);
    if (!$im) {
        return;
    }
    imagepng($im);
}

function width(string $bytes): int
{
    $im = imagecreatefromstring($bytes);
    if ($im) {
        return imagesx($im);
    }
    return 0;
}

function height(string $bytes): int
{
    $im = imagecreatefromstring($bytes);
    return $im === false ? 0 : imagesy($im);
}

function typed(string $bytes): int
{
    $im = imagecreatefromstring($bytes);
    if (!$im instanceof GdImage) {
        throw new RuntimeException('Invalid image');
    }
    return imagesx($im);
}

function inline(string $bytes): int
{
    if (($im = imagecreatefromstring($bytes)) === false) {
        throw new RuntimeException('Invalid image');
    }
    return imagesx($im);
}

function strictOrStop(string $bytes): void
{
    $im = imagecreatefromstring($bytes) or die('Invalid image');
    imagegif($im);
}

function asserted(string $bytes): void
{
    $im = imagecreatefromstring($bytes);
    assert($im !== false);
    imagejpeg($im);
}

function shortCircuit(string $bytes): void
{
    $im = imagecreatefromstring($bytes);
    $im !== false && imagepng($im);
}

function each(array $blobs): void
{
    foreach ($blobs as $bytes) {
        $im = imagecreatefromstring($bytes);
        if (empty($im)) {
            continue;
        }
        imagepng($im);
    }
}
