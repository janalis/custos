<?php
function purge(array $cache, $tmp, $lock) {
    unset($cache['k'] /* stale entry */, $cache['j'] // gone too
    );
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($tmp);</weak_warning>
    $tmp = 1;
    unset($tmp);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($lock /* held */);</weak_warning>
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($cache);</weak_warning>
}
$a = 1; $b = 2; $c = 3; $d = 4;
unset($a); <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($b) ?></weak_warning><p>html</p>
<?php unset($c) ?>
<?php <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($d) ?></weak_warning><p>done</p>
