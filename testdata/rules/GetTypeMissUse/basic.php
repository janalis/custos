<?php
function kinds($item, $mode = 'array') {
    $a = <warning descr="Use 'is_int($item->qty)' instead.">gettype($item->qty) === 'integer'</warning>;
    $b = <warning descr="Use '!is_bool($item)' instead.">'boolean' !== gettype($item)</warning>;
    $c = <warning descr="Use 'is_null($item)' instead.">gettype($item) == 'NULL'</warning>;
    $d = <warning descr="Use '!is_float($item)' instead.">gettype($item) != 'double'</warning>;
    $e = <warning descr="Use 'is_array($item)' instead.">gettype($item) === $mode</warning>;
    $f = gettype($item) === <error descr="gettype() never returns 'int'.">'int'</error>;
    $g = gettype($item) === <error descr="gettype() never returns 'null'.">'null'</error>;

    $h = gettype($item) === 'unknown type';
    $i = gettype($item) !== 'resource (closed)';
    $j = (gettype($item)) === 'string';
    $k = gettype($item) === ($item ? 'string' : 'object');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
