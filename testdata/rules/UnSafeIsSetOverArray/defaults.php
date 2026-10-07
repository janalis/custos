<?php
function f($node, array $list, $k) {
    if (isset(<warning descr="Compute the concatenated key in a variable before using it.">$list['p' . $k]</warning>)) {}
    $stored = isset($list['p' . $k]);
    $nested = isset($list['p' . $k]['x']);
    $a = isset($node);
    $b = isset($list['k']);
    return [$stored, $nested, $a, $b];
}
