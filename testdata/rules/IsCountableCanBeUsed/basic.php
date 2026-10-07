<?php
function sizeOf($bag, $other) {
    $a = <weak_warning descr="Use 'is_countable($bag)' instead of the is_array()/instanceof Countable pair.">is_array($bag)</weak_warning> || $bag instanceof \Countable;
    $b = ($bag instanceof Countable) || <weak_warning descr="Use 'is_countable($bag->items)' instead of the is_array()/instanceof Countable pair.">is_array($bag->items)</weak_warning> || $bag->items instanceof Countable;
    $c = $other === null || (<weak_warning descr="Use 'is_countable($other)' instead of the is_array()/instanceof Countable pair.">is_array($other)</weak_warning> || ($other instanceof \Countable));
    $d = $bag instanceof Some\Countable || \is_array($bag); // Some\Countable is not \Countable
}
