<?php
function keyed($s, $t, $u, $k) {
    $s = str_replace(['amp' => '&', 'lt' => '<'], ['and', 'less'], $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['amp' => '@', 'gt' => '>'], ['at', 'more'], $s)</warning>;

    $t = str_replace([1 => 'a', 'z'], 'x', $t);
    $t = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['b', 'c'], 'y', $t)</warning>;

    $u = str_replace([$k => 'a', 'z'], 'x', $u);
    $u = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('b', 'y', $u)</warning>;

    return str_replace(['0' => 'q', 0 => 'r'], 'z', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace('p', 'w', $s . $t . $u)</warning>);
}

function plain($v) {
    $v = str_replace(['a', 'b'], ['1', '2'], $v);
    $v = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c'], ['3'], $v)</warning>;
    return $v;
}
