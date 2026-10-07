<?php
function kinds($item, $mode = 'array') {
    $a = is_int($item->qty);
    $b = !is_bool($item);
    $c = is_null($item);
    $d = !is_float($item);
    $e = is_array($item);
    $f = gettype($item) === 'int';
    $g = gettype($item) === 'null';

    $h = gettype($item) === 'unknown type';
    $i = gettype($item) !== 'resource (closed)';
    $j = (gettype($item)) === 'string';
    $k = gettype($item) === ($item ? 'string' : 'object');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
