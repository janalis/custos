<?php
function contexts($row, $ok, array $rest)
{
    $a = 'v=' . <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>;
    $b = <weak_warning descr="Replace with '(include 'x.php') === null'.">is_null(include 'x.php')</weak_warning>;
    $c = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning> ? 1 : 2;
    $d = $ok ?? <weak_warning descr="Replace with '$row !== null'.">!is_null($row)</weak_warning>;
    $e = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning> == $ok;
    $f = is_null(...$rest);
    $g = is_null(...);
    $h = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning> === PHP_EOL;
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}
