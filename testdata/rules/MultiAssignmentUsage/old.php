<?php
foreach ($pairs as $idx => $pair) {
    list($left, $right) = $pair;
    if ($idx > 0) {
        list($k) = $idx;
    }
    list($m, $n) = $other;                // E2: not a loop variable
    $fn = function () use ($pair) { list($u) = $pair; };   // E2: closure boundary

    $width = $pair[0];
    // a plain comment
    /** @var int $height */
    <weak_warning descr="Use one destructuring assignment from '$pair' instead.">$height = $pair[1]</weak_warning>;
    <weak_warning descr="Use one destructuring assignment from '$pair' instead.">$depth = $pair[-1]</weak_warning>;

    $name = $pair['name'];
    $kind = $pair['kind'];                // E4
    $copy = array();
    $copy[0] = $pair[1];
    $copy[1] = $pair[0];                  // E5
}
