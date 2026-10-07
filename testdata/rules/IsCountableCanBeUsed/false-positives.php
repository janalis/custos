<?php
function sizeOf($bag, $other) {
    $d = is_array($bag) || $other instanceof Countable;           // different subject
    $e = (is_array($bag)) || $bag instanceof Countable;           // call is parenthesised
    $f = is_array($bag) || $bag instanceof Traversable;           // other interface
    $g = is_array($bag) || ($bag instanceof Countable && $other); // inside &&
    $h = is_array($bag) or $bag instanceof Countable;             // keyword or
    $i = is_array($bag) && $x || $bag instanceof Countable;       // direct parent is &&
    $j = !is_array($bag) || $bag instanceof Countable;            // negated
    $k = is_array($bag, 1) || $bag instanceof Countable;          // argument count
}
