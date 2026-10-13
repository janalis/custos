<?php
function unrelatedCounter(array $hits): float
{
    $x = 2;
    $other = 0;
    $other++;
    $hits[0]++;
    $read = function () use ($x, &$other) {
        return $x + $other;
    };
    $read();
    return <warning descr="Keep the inverse trigonometric argument within minus one and one.">asin($x)</warning>;
}
function steppedAfterCall(): float
{
    $x = 2;
    $r = <warning descr="Keep the inverse trigonometric argument within minus one and one.">acos($x)</warning>;
    $x--;
    return $r;
}
function steppedInNestedFunction(): float
{
    $x = 2;
    $inner = function () {
        $x = 5;
        $x--;
        extract([]);
        return function () use (&$x) {
            return $x;
        };
    };
    return <warning descr="Keep the inverse trigonometric argument within minus one and one.">asin($x)</warning> + $inner()();
}
function helperCounter(): int
{
    $top = 0;
    $top++;
    return $top;
}
$top = -3;
echo <warning descr="Keep the inverse trigonometric argument within minus one and one.">asin($top)</warning>;
