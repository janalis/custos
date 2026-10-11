<?php
function walkAll($bag, $other) {
    $a = is_iterable($bag);
    $b = ($bag instanceof Traversable) || is_array($bag->items) || $bag->items instanceof Traversable;
    $c = $other === null || (is_iterable($other));
    $d = \is_array($bag[0]) || $bag [0] instanceof Some\Traversable; // Some\Traversable is not \Traversable
    return [$a, $b, $c, $d];
}
