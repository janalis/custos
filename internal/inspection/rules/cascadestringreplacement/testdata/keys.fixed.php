<?php
function keyed($s, $t, $u, $k) {
    $s = str_replace(['&', '<', '@', '>'], ['and', 'less', 'at', 'more'], $s);

    $t = str_replace(['a', 'z', 'b', 'c'], ['x', 'x', 'y', 'y'], $t);

    $u = str_replace([$k => 'a', 'z'], 'x', $u);
    $u = str_replace('b', 'y', $u);

    return str_replace(['0' => 'q', 0 => 'r'], 'z', str_replace('p', 'w', $s . $t . $u));
}

function plain($v) {
    $v = str_replace(['a', 'b', 'c'], ['1', '2', '3'], $v);
    return $v;
}
