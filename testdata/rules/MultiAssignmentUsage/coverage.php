<?php
function cover(array $x, $o, array $rows) {
    $a .= $x[0];
    $b = $x[1];                       // previous assignment is compound
    <weak_warning descr="Use one destructuring assignment from '$x' instead.">$c = $x[010]</weak_warning>;
    <weak_warning descr="Use one destructuring assignment from '$x' instead.">$d = $x[9]</weak_warning>;   // 010 is key 8
    $e = $x[1e300];
    $f = $x[1];                       // key out of range
    $g = $o -> p[0];
    <weak_warning descr="Use one destructuring assignment from '$o->p' instead.">$h = $o->p[1]</weak_warning>;
    foreach ($rows as [, $k]) {
        [$m, $n] = $other;            // header does not declare $other
    }
}
