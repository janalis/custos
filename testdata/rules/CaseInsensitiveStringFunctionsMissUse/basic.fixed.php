<?php
function parts($path, $mode)
{
    $sep = '::';
    $a = strpos($path, '/');
    $b = strrpos($path, '.', -2);
    $c = \strstr($path, $sep);
    $d = strpos($path, "2024");
    return [$a, $b, $c, $d];
}
