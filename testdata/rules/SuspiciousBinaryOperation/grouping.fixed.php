<?php
function asInt(?int $limit) {
    return (int) ($limit >= 10);
}

function twice(?int $limit, $other) {
    return !($limit > 10) && $other;
}

function equals(?int $limit, bool $want, int $step) {
    $a = ($limit >= 1.5) == $want;
    $b = ($limit >= $step = 2) && $want;
    return $a ? $b : $limit >= $step;
}

function unordered(?float $ratio, ?array $left, array $right, ?int $count, float $max) {
    $a = !($ratio < 10);
    $b = !($left < $right);
    return $a && $b && !($count <= $max);
}

function tooShort(string $name, int $min, $strict) {
    if (!(strlen($name) >= 2)) {}
    if (!(strlen($name) >= 2)) {}
    if ((strlen($name) >= $min = 3) && $strict) {}
}
