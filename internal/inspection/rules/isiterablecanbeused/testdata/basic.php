<?php
function walkAll($bag, $other) {
    $a = <weak_warning descr="Use 'is_iterable($bag)' instead of the is_array()/instanceof Traversable pair.">is_array($bag)</weak_warning> || $bag instanceof \Traversable;
    $b = ($bag instanceof Traversable) || <weak_warning descr="Use 'is_iterable($bag->items)' instead of the is_array()/instanceof Traversable pair.">is_array($bag->items)</weak_warning> || $bag->items instanceof Traversable;
    $c = $other === null || (<weak_warning descr="Use 'is_iterable($other)' instead of the is_array()/instanceof Traversable pair.">is_array($other)</weak_warning> || ($other instanceof \Traversable));
    $d = \is_array($bag[0]) || $bag [0] instanceof Some\Traversable; // Some\Traversable is not \Traversable
    return [$a, $b, $c, $d];
}
