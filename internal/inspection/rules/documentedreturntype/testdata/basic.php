<?php
declare(strict_types=1);

namespace Shop;

class Order {}
final class Coupon {}

// The @return type rejects the value.
/** @return Order */
function latest()
{
    return <warning descr="Return Order from Shop\latest(), not string.">'order-1'</warning>;
}

/** @return Order|null */
function find(): ?object
{
    return <warning descr="Return Order|null from Shop\find(), not Shop\Coupon.">new Coupon()</warning>;
}

/** @return list<Order> */
function recent()
{
    return <warning descr="Return list<Order> from Shop\recent(), not string.">'none'</warning>;
}

/** @phpstan-return int */
function count_all()
{
    return <warning descr="Return int from Shop\count_all(), not array.">[]</warning>;
}

// The native type rejects a value only PHPDoc describes.
/** @param string $amount */
function cents($amount): int
{
    return <warning descr="Return int from Shop\cents(), not string.">$amount</warning>;
}

class Cart
{
    /** @return $this */
    public function touch(): object
    {
        return <warning descr="Return $this from Shop\Cart::touch(), not Shop\Coupon.">new Coupon()</warning>;
    }
}

/** @phpstan-type LineId int */
class Ledger
{
    /** @return LineId */
    public function last()
    {
        return <warning descr="Return LineId from Shop\Ledger::last(), not string.">'line-1'</warning>;
    }
}

$total = /** @return int */ fn() => <warning descr="Return int from the arrow function, not string.">'1'</warning>;

// Value details count against exact values.
/** @return negative-int */
function offset()
{
    return <warning descr="Return negative-int from Shop\offset(), not int.">count([1, 2])</warning>;
}

/** @return positive-int */
function stock(bool $empty)
{
    if ($empty) {
        return <warning descr="Return positive-int from Shop\stock(), not int.">0</warning>;
    }
    return <warning descr="Return positive-int from Shop\stock(), not int.">-3</warning>;
}

/** @return int<1, 9> */
function digit(bool $high)
{
    return <warning descr="Return int<1,9> from Shop\digit(), not int.">$high ? 10 : 12</warning>;
}

/** @return non-empty-string */
function label()
{
    return <warning descr="Return non-empty-string from Shop\label(), not string.">''</warning>;
}

