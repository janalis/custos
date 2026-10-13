<?php
function lookup(object $o, ?object $maybe, array $a, object|array $either) {
    $x = isset($o['k']);
    $y = isset($maybe['k']);
    $z = isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$a['k']</weak_warning>);
    $w = isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$either['k']</weak_warning>);
    return [$x, $y, $z, $w];
}
