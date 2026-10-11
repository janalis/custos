<?php
<error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv(PHP_INT_MIN, -1)</error>;

function plainOperands(): int
{
    $min = PHP_INT_MIN;
    $divisor = -1;

    return <error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv($min, $divisor)</error>;
}

function copiedOperand(): int
{
    $source = PHP_INT_MIN;
    $min = $source;

    return <error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv($min, -1)</error>;
}

function namedOperands(): int
{
    $min = PHP_INT_MIN;

    return <error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv(num2: -1, num1: $min)</error>;
}

function readByValue(): int
{
    $min = PHP_INT_MIN;
    $size = abs($min) + max(...[1, 2]) + strlen(string: 'x');
    $pair = [$min, $size];
    $text = sprintf('%d %d %d', 1, 2, $min);
    $next = $min + 1;
    $copy = $min;
    $show = function () use ($min) {
        return $min;
    };

    return <error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv($min, -1)</error> + count($pair) + $copy + $show() + $next + strlen($text);
}

function changedAfterwards(): int
{
    $min = PHP_INT_MIN;
    $result = <error descr="Avoid dividing PHP_INT_MIN by negative one.">intdiv($min, -1)</error>;
    $min++;

    return $result;
}
