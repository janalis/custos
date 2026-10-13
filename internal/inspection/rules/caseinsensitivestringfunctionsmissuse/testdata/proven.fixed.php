<?php
const SEP = '/';
const WIDTH = 5;
define('OTHER', 'x');

class Splitter
{
    const GLUE = '::';

    public function cut($text)
    {
        return strstr($text, self::GLUE);
    }
}

function proven($path)
{
    $dot = '.';
    $parts = explode($dot, $path);
    $a = strrpos($path, $dot);
    $b = strpos($path, SEP);
    $c = strpos($path, ('#'));
    $d = stripos($path, WIDTH);
    return [$parts, $a, $b, $c, $d];
}
