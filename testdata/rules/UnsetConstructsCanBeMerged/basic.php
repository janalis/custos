<?php
function release(array $cache, $tmp, $lock) {
    unset($cache['k1']);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($tmp);</weak_warning>
    // the lock goes too
    /** trailing note */
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset ( $lock );</weak_warning>

    $tmp = 1;
    unset($tmp, $cache);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($lock);</weak_warning>

    if ($tmp) {
        unset ($cache) ;
        <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($a, $b);</weak_warning>
    }
    unset($tmp);
}
