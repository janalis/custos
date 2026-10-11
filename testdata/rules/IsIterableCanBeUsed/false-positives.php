<?php
function pull(array &$queue) { return array_pop($queue); }

function walkAll($bag, $other, $x, $i, $o) {
    $d = is_array($bag) || $other instanceof Traversable;           // different subject
    $e = (is_array($bag)) || $bag instanceof Traversable;           // call is parenthesised
    $f = is_array($bag) || $bag instanceof Iterator;                // other interface
    $g = is_array($bag) || ($bag instanceof Traversable && $other); // inside &&
    $h = is_array($bag) or $bag instanceof Traversable;             // keyword or
    $i = is_array($bag) && $x || $bag instanceof Traversable;       // direct parent is &&
    $j = !is_array($bag) || $bag instanceof Traversable;            // negated
    $k = is_array($bag, $x) || $bag instanceof Traversable;         // two arguments
    // Evaluating the operand once instead of twice would change the result.
    $l = is_array($x[$i++]) || $x[$i++] instanceof Traversable;     // increment
    $m = is_array(next($x)) || next($x) instanceof Traversable;     // moves the array pointer
    $n = is_array($o = $x) || ($o = $x) instanceof Traversable;     // assignment
    $p = is_array(pull($x)) || pull($x) instanceof Traversable;     // by-reference parameter
    $q = is_array($o->read(fgets($x))) || $o->read(fgets($x)) instanceof Traversable; // nested stream read
    return [$l, $m, $n, $p, $q, $d, $e, $f, $g, $h, $i, $j, $k];
}
