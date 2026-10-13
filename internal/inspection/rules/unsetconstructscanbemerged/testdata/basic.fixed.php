<?php
function release(array $cache, $tmp, $lock) {
    unset($cache['k1'], $tmp, $lock);
    // the lock goes too
    /** trailing note */

    $tmp = 1;
    unset($tmp, $cache, $lock);

    if ($tmp) {
        unset($cache, $a, $b);
    }
    unset($tmp);
}
