<?php
// Split assignments in a brace-less control body get braces.
function f($c, array $xs)
{
    if ($c) {
        $_SESSION['a'] = 10;
        $n = 10;
    }
    else $n = 2;
    foreach ($xs as $x) {
        $o = $x;
        $m = $x;
    }
    declare(ticks=1) {
        $q = 3;
        $p = 3;
    }
    return [$n, $m, $o, $p, $q];
}
