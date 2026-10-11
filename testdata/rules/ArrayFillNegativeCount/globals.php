<?php

$width = -1;
$other = -1;
$viaSuperglobal = -1;
resize();
array_fill(0, $width, 'c');
array_fill(0, $viaSuperglobal, 'c');

function resize()
{
    global $width, $$dynamic;
    $width = 5;
    $GLOBALS['viaSuperglobal'] = 5;
}
