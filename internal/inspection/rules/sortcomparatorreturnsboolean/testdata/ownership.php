<?php
// Inline literals with no return type, or int: the call owns the comparator.
<warning descr="Return a three-way integer comparison result.">usort($xs, function ($a, $b) { return $a < $b; })</warning>;
<warning descr="Return a three-way integer comparison result.">uasort($xs, (static fn(int $a, int $b): INT => $a > $b))</warning>;
<warning descr="Return a three-way integer comparison result.">uksort(callback: fn($a, $b) => ($a > $b), array: $xs)</warning>;
// Declared non-int return types: an integer would be coerced or rejected.
<warning descr="Return a three-way integer comparison result.">usort($xs, function (int $a, int $b): bool { return $a > $b; })</warning>;
<warning descr="Return a three-way integer comparison result.">usort($xs, fn(int $a, int $b): bool => $a < $b)</warning>;
<warning descr="Return a three-way integer comparison result.">usort($xs, fn($a, $b): ?int => $a > $b)</warning>;
<warning descr="Return a three-way integer comparison result.">usort($xs, fn($a, $b): mixed => $a > $b)</warning>;
// Comparators reachable from other callers keep their boolean result.
function isNewer($a, $b) { return $a > $b; }
<warning descr="Return a three-way integer comparison result.">usort($versions, 'isNewer')</warning>;
echo isNewer(1, 2) ? "newer" : "older";
$gt = fn($a, $b) => $a > $b;
<warning descr="Return a three-way integer comparison result.">usort($versions, $gt)</warning>;
var_dump($gt(1, 2));
