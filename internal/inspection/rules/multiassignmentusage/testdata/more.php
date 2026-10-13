<?php
function f(array $rows) {
    foreach ($rows as [$a, list($b, $c)]) {
        <weak_warning descr="Destructure directly in the foreach header.">[$x, $y] = $c</weak_warning>;
        foreach ($b as $inner) {
            list($p) = $a;                // E2b: the inner loop does not declare $a
        }
    }
    $one = load()[0];
    $two = load( )[1];                    // E6b: load() would run once instead of twice
    $first = $rows[0];
    $again = $rows[0];                    // E7: same key
    <weak_warning descr="Use one destructuring assignment from '$rows' instead.">$f = $rows[1.5]</weak_warning>;
}
