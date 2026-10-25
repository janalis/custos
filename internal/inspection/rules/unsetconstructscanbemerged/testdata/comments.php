<?php
function drop(array $ctx, $tmp, $lock) {
    unset($ctx['a']);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($ctx['b']);</weak_warning> // grows with every call
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($ctx['c']);</weak_warning> // unique per call
    $tmp = 1;
    unset($tmp);
    // the lock goes too
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($lock);</weak_warning>
    $tmp = 2;
    unset($ctx['d']);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($ctx['e']);</weak_warning> // both are derived
}
