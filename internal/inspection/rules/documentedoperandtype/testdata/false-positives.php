<?php

class Money extends GMP {}

class Orphan extends MissingBase {}

/**
 * @param array $a
 * @param array $b
 * @param array|int $either
 * @param mixed $anything
 * @param int $number
 * @param Money $money
 * @param Orphan $orphan
 */
function documented($a, $b, $either, $anything, $number, $money, $orphan, $unknown)
{
    $x = $a + $b;
    $x = $a + $unknown;
    $x = $a + $anything;
    $x = $either * 2;
    $x = $anything * 2;
    $x = $number + 1;
    $x = $money * 2;
    $x = $orphan + 1;
}

function proven(array $row)
{
    return $row * 2;
}

function vivified(int $n)
{
    $list = false;
    $list[] = 1;
    return $n;
}

class Rebound
{
    public function bind()
    {
        return function () {
            return $this + 1;
        };
    }
}

class Plain {}

function contradicted()
{
    /** @var Plain $amount */
    $amount = 5;
    return $amount + 1;
}
