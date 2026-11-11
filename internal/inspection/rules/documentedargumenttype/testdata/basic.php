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

// The native type rejects a value only PHPDoc describes.
/** @param string $amount */
function bill($amount)
{
    charge(<warning descr="Pass int to $cents of Shop\charge(), not string.">$amount</warning>);
}
