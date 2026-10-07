<?php
function f($node, array $list) {
    $a = null !== $node;
    $b = null === $node;
    $c = isset($list['k']);
    return [$a, $b, $c];
}
