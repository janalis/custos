<?php
function contexts($row, $ok, array $rest)
{
    $a = 'v=' . ($row === null);
    $b = (include 'x.php') === null;
    $c = $row === null ? 1 : 2;
    $d = $ok ?? $row !== null;
    $e = ($row === null) == $ok;
    $f = is_null(...$rest);
    $g = is_null(...);
    $h = ($row === null) === PHP_EOL;
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}
