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

// The native type rejects a value only PHPDoc describes.
/** @param string $amount */
function bill($amount)
{
    charge(<warning descr="Pass int to $cents of Shop\charge(), not string.">$amount</warning>);
}
