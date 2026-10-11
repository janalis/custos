<?php
const SEP = '/';
const WIDTH = 5;
define('OTHER', 'x');

class Splitter
{
    const GLUE = '::';

    public function cut($text)
    {
        return <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">stristr($text, self::GLUE)</weak_warning>;
    }
}

function proven($path)
{
    $dot = '.';
    $parts = explode($dot, $path);
    $a = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">strripos($path, $dot)</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, SEP)</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, ('#'))</weak_warning>;
    $d = stripos($path, WIDTH);
    return [$parts, $a, $b, $c, $d];
}
