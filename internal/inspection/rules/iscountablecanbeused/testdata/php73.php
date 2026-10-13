<?php
function sizeOf($bag, $other) {
    $a = is_array($bag) || $bag instanceof \Countable;
    $b = ($bag instanceof Countable) || is_array($bag->items) || $bag->items instanceof Countable;
    $c = $other === null || (is_array($other) || ($other instanceof \Countable));
    $d = $bag instanceof Some\Countable || \is_array($bag);
}
