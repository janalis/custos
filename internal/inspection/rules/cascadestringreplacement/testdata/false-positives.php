<?php
function f($a, $o, $line) {
    $a = str_replace('x', 'y', $a);
    $b = str_replace('z', 'w', $a);
    $a = str_replace('x', 'y', $a);
    echo $a;
    $a = str_replace('z', 'w', $a);
    $o->p = str_replace('x', 'y', $o->p);
    $o->p = str_replace('z', 'w', $o->p);
    $c = str_replace(['a', 'b'], 'c', $line);
    $d = str_replace(['a', 'a'], ['c'], $line);
    $e = str_replace([], 'c', $line);
    $h = str_replace(['a', "a"], 'c', $line);
    $i = str_replace(['a', 1], 'c', $line);
    $line .= str_replace('k', 'l', $line);
    return [$b, $c, $d, $e, $g, $h, $i];
}

// The second call's arguments read the variable after the first replacement.
function reads_subject($s)
{
    $s = str_replace('a', 'b', $s);
    $s = str_replace('x', strlen($s), $s);
    $s = str_replace(substr($s, 0, 1), 'y', $s);
    return $s;
}

// Folding into the return would drop a write the caller still sees.
function by_reference(&$s)
{
    $s = str_replace('a', 'b', $s);
    return str_replace('c', 'd', $s);
}

function via_global()
{
    global $shared;
    $shared = str_replace('a', 'b', $shared);
    return str_replace('c', 'd', $shared);
}

function via_static()
{
    static $acc = 'ac';
    $acc = str_replace('a', 'b', $acc);
    return str_replace('c', 'd', $acc);
}

function via_closure($v)
{
    return function () use (&$v) {
        $v = str_replace('b', 'z', $v);
        return str_replace('c', 'd', $v);
    };
}
