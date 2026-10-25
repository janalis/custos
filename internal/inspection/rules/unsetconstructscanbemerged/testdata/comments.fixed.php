<?php
function drop(array $ctx, $tmp, $lock) {
    unset($ctx['a']);
    unset($ctx['b']); // grows with every call
    unset($ctx['c']); // unique per call
    $tmp = 1;
    unset($tmp, $lock);
    // the lock goes too
    $tmp = 2;
    unset($ctx['d'], $ctx['e']); // both are derived
}
