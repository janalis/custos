<?php
// Inline literals with no return type, or int: the call owns the comparator.
usort($xs, function ($a, $b) { return $b <=> $a; });
uasort($xs, (static fn(int $a, int $b): INT => $a <=> $b));
uksort(callback: fn($a, $b) => ($a <=> $b), array: $xs);
// Declared non-int return types: an integer would be coerced or rejected.
usort($xs, function (int $a, int $b): bool { return $a > $b; });
usort($xs, fn(int $a, int $b): bool => $a < $b);
usort($xs, fn($a, $b): ?int => $a > $b);
usort($xs, fn($a, $b): mixed => $a > $b);
// Comparators reachable from other callers keep their boolean result.
function isNewer($a, $b) { return $a > $b; }
usort($versions, 'isNewer');
echo isNewer(1, 2) ? "newer" : "older";
$gt = fn($a, $b) => $a > $b;
usort($versions, $gt);
var_dump($gt(1, 2));
