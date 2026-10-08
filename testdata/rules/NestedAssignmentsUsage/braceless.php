<?php
// Split assignments in a brace-less control body get braces.
function f($c, array $xs)
{
    if ($c) <weak_warning descr="Split this chained assignment into separate assignments.">$n = $_SESSION['a'] = 10</weak_warning>;
    else $n = 2;
    foreach ($xs as $x) <weak_warning descr="Split this chained assignment into separate assignments.">$m = $o = $x</weak_warning>;
    declare(ticks=1) <weak_warning descr="Split this chained assignment into separate assignments.">$p = $q = 3</weak_warning>;
    return [$n, $m, $o, $p, $q];
}
