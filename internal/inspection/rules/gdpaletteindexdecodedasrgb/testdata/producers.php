<?php
// The index and its image must be what every path agrees on.

function onePath($flag)
{
    $im = imagecreate(3, 3);
    $c = imagecolorallocate($im, 255, 0, 0);

    return <warning descr="Look up palette components instead of shifting the index.">($c >> 16) & 255</warning>;
}

function packedColourOnOnePath($flag)
{
    $im = imagecreate(3, 3);
    if ($flag) {
        $c = imagecolorallocate($im, 255, 0, 0);
    } else {
        $c = 0xff0000;
    }

    return ($c >> 16) & 255;
}

function otherImageOnOnePath($flag, $other)
{
    $im = imagecreate(3, 3);
    if ($flag) {
        $c = imagecolorallocate($im, 255, 0, 0);
    } else {
        $c = imagecolorallocate($other, 255, 0, 0);
    }

    return ($c >> 16) & 255;
}

function truecolorOnOnePath($flag)
{
    if ($flag) {
        $im = imagecreate(3, 3);
    } else {
        $im = imagecreatetruecolor(3, 3);
    }
    $c = imagecolorallocate($im, 255, 0, 0);

    return ($c >> 16) & 255;
}

function truecolorOnTheFirstPath($flag)
{
    if ($flag) {
        $im = imagecreatetruecolor(3, 3);
    } else {
        $im = imagecreate(3, 3);
    }
    $c = imagecolorallocate($im, 255, 0, 0);

    return ($c >> 16) & 255;
}

// The shift only runs for the truecolor image.
function decodedForTheTruecolorImage($flag)
{
    if ($flag) {
        $im = imagecreate(3, 3);
    } else {
        $im = imagecreatetruecolor(3, 3);
    }
    $c = imagecolorallocate($im, 255, 0, 0);
    if (!$flag) {
        return ($c >> 16) & 255;
    }

    return imagecolorsforindex($im, $c)['red'];
}

function packedColourOnALaterPass($count)
{
    $im = imagecreate(3, 3);
    $c = imagecolorallocate($im, 255, 0, 0);
    while ($count--) {
        echo ($c >> 16) & 255;
        $c = 0xff0000;
    }
}

function otherImageOnALaterPass($count, $other)
{
    $im = imagecreate(3, 3);
    while ($count--) {
        $c = imagecolorallocate($im, 255, 0, 0);
        echo ($c >> 16) & 255;
        $im = $other;
    }
}
