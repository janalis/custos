<?php
function f($node, array $list) {
    $a = <weak_warning descr="Compare with null instead: 'null !== $node'.">isset($node)</weak_warning>;
    $b = <weak_warning descr="Compare with null instead: 'null === $node'.">!isset($node)</weak_warning>;
    $c = isset($list['k']);
    return [$a, $b, $c];
}
