<?php
function asInt(?int $limit) {
    return (int) <error descr="Null or false operands make this negated comparison misleading; use '$limit >= 10'.">!($limit < 10)</error>;
}

function twice(?int $limit, $other) {
    return !<error descr="Null or false operands make this negated comparison misleading; use '$limit > 10'.">!($limit <= 10)</error> && $other;
}

function equals(?int $limit, bool $want, int $step) {
    $a = <error descr="Null or false operands make this negated comparison misleading; use '$limit >= 1.5'.">!($limit < 1.5)</error> == $want;
    $b = <error descr="Null or false operands make this negated comparison misleading; use '$limit >= $step = 2'.">!($limit < $step = 2)</error> && $want;
    return $a ? $b : <error descr="Null or false operands make this negated comparison misleading; use '$limit >= $step'.">!($limit < $step)</error>;
}

function unordered(?float $ratio, ?array $left, array $right, ?int $count, float $max) {
    $a = <error descr="Null or false operands make this negated comparison misleading; use '$ratio >= 10'.">!($ratio < 10)</error>;
    $b = <error descr="Null or false operands make this negated comparison misleading; use '$left >= $right'.">!($left < $right)</error>;
    return $a && $b && <error descr="Null or false operands make this negated comparison misleading; use '$count > $max'.">!($count <= $max)</error>;
}

function tooShort(string $name, int $min, $strict) {
    if (!strlen($name <error descr="This comparison probably belongs outside the call parentheses.">>=</error> 2)) {}
    if (!(strlen($name <error descr="This comparison probably belongs outside the call parentheses.">>=</error> 2))) {}
    if (strlen($name <error descr="This comparison probably belongs outside the call parentheses.">>=</error> $min = 3) && $strict) {}
}
