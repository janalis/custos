<?php
function f($a, $b, DateTime $d, \PDO $pdo) {
    return [
        $a === $b,
        $a !== 'x',
        $a < 'x',
        $d == $a,
        $pdo != $b,
    ];
}

// Only DateTime and DateTimeImmutable implement DateTimeInterface: both
// compare their dates with ==.
function sameDate(mixed $value, \Closure $get, \DateTimeInterface $a, $b): bool
{
    return $value instanceof \DateTimeInterface && $value == $get() || $a == $b;
}
