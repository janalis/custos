<?php
<error descr="Supply a nonzero integer divisor.">intdiv(14, 0)</error>;

$zero = 0;
<error descr="Supply a nonzero integer divisor.">intdiv(14, $zero)</error>;
$zero++;

function countedAfterwards(array $rows): int
{
    $seen = 0;
    $share = <error descr="Supply a nonzero integer divisor.">intdiv(100, $seen)</error>;
    foreach ($rows as [$left, $right]) {
        $seen++;
    }
    $bump = function () use (&$seen) {
        $seen = 2;
    };
    $bump();

    return $share;
}

function unrelatedCounter(array $rows): int
{
    $parts = 0;
    $index = 0;
    foreach ($rows as $key => $row) {
        $index++;
    }
    try {
        $index--;
    } catch (RuntimeException $problem) {
        $index = 1;
    } catch (LogicException) {
        $index = 2;
    }
    $reset = function () use (&$index, $parts) {
        $index = 0;
    };

    return <error descr="Supply a nonzero integer divisor.">intdiv($index, ($parts))</error>;
}
