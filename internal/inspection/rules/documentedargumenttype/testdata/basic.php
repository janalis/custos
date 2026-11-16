<?php
declare(strict_types=1);

namespace Shop;

class Order {}
final class Coupon {}

/** @param Order $order */
function archive($order) {}

/** @param list<Order> $orders */
function batch(array $orders) {}

/** @param Order|null $order */
function maybeArchive(?object $order) {}

function charge(int $cents) {}

// The PHPDoc type rejects the value.
archive(<warning descr="Pass Shop\Order to $order of Shop\archive(), not string.">'order-1'</warning>);
archive(<warning descr="Pass Shop\Order to $order of Shop\archive(), not Shop\Coupon.">new Coupon()</warning>);
maybeArchive(<warning descr="Pass Shop\Order|null to $order of Shop\maybeArchive(), not Shop\Coupon.">new Coupon()</warning>);

// Ranges and lengths count against exact values.
/** @param positive-int $quantity */
function restock(int $quantity) {}
restock(<warning descr="Pass positive-int to $quantity of Shop\restock(), not 0.">0</warning>);
restock(<warning descr="Pass positive-int to $quantity of Shop\restock(), not -2.">-2</warning>);

/** @param int<1, 9>|null $digit */
function dial($digit = null) {}
dial(<warning descr="Pass int<1, 9>|null to $digit of Shop\dial(), not 10.">10</warning>);

/** @param non-empty-string $code */
function redeem(string $code) {}
redeem(<warning descr="Pass non-empty-string to $code of Shop\redeem(), not ''.">''</warning>);

// Literal string values count against exact values.
/** @param 'asc'|'desc' $direction */
function sortBy(string $direction) {}
sortBy(<warning descr="Pass 'asc'|'desc' to $direction of Shop\sortBy(), not 'up'.">'up'</warning>);
function descending()
{
    $direction = 'down';
    sortBy(<warning descr="Pass 'asc'|'desc' to $direction of Shop\sortBy(), not 'down'.">$direction</warning>);
}

// The native type rejects a value only PHPDoc describes.
/** @param string $amount */
function bill($amount)
{
    charge(<warning descr="Pass int to $cents of Shop\charge(), not string.">$amount</warning>);
}

// A write into false makes it an array (deprecated from PHP 8.1): the
// array is still rejected.
/** @param int $count */
function countTo($count) {}
function vivifiedArg(): void
{
    $list = false;
    $list[] = 1;
    countTo(<warning descr="Pass int to $count of Shop\countTo(), not int[].">$list</warning>);
}

// Also into a property only PHPDoc types, or declared mixed.
class DocLazyCount
{
    /** @var array|false */
    private $items = false;
    /** @var array|false */
    private mixed $any = false;

    public function fill(): void
    {
        if ($this->items === false) {
            $this->items[] = 1;
            countTo(<warning descr="Pass int to $count of Shop\countTo(), not array.">$this->items</warning>);
        }
        if ($this->any === false) {
            $this->any[] = 1;
            countTo(<warning descr="Pass int to $count of Shop\countTo(), not array.">$this->any</warning>);
        }
    }
}
