<?php
function flags(array $list, $x, $y) {
    $list[] = ($x or $y);
    $list[] = ($x and $y);
    $list[] = ($x xor $y);
    $list[] = $x || $y;
    $list[] = $x ? 1 : 2;
    return $list;
}
