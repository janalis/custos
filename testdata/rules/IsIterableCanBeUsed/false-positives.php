<?php
function walkAll($bag, $other, $x) {
    $d = is_array($bag) || $other instanceof Traversable;           // different subject
    $e = (is_array($bag)) || $bag instanceof Traversable;           // call is parenthesised
    $f = is_array($bag) || $bag instanceof Iterator;                // other interface
    $g = is_array($bag) || ($bag instanceof Traversable && $other); // inside &&
    $h = is_array($bag) or $bag instanceof Traversable;             // keyword or
    $i = is_array($bag) && $x || $bag instanceof Traversable;       // direct parent is &&
    $j = !is_array($bag) || $bag instanceof Traversable;            // negated
    $k = is_array($bag, $x) || $bag instanceof Traversable;         // two arguments
    return [$d, $e, $f, $g, $h, $i, $j, $k];
}
