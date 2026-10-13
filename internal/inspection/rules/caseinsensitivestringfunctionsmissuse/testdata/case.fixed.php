<?php
function split_path($path)
{
    $a = strpos($path, '/');
    $b = \strrpos($path, '.');
    $c = strstr($path, '::');
    $d = STRIPOS($path, 'x');
    return [$a, $b, $c, $d];
}
