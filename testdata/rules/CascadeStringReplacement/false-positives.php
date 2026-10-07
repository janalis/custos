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
