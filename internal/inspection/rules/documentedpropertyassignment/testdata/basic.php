<?php
declare(strict_types=1);

namespace Shop;

class Order {}
final class Coupon {}

final class Cart
{
    /** @var int */
    public $count = 0;
    /** @var Order|null */
    public $order;
    /** @var list<Order> */
    public array $orders = [];
    public int $total = 0;
    /** @var string */
    public static $currency = 'EUR';
    /** @var positive-int */
    public $page = 1;
    /** @var non-empty-string */
    public string $label = 'cart';

    /** @param string $amount */
    public function fill($amount, Coupon $coupon)
    {
        // The PHPDoc type rejects the value.
        $this->count = <warning descr="Assign int to Shop\Cart::$count, not string.">'3'</warning>;
        $this->order = <warning descr="Assign Shop\Order|null to Shop\Cart::$order, not Shop\Coupon.">$coupon</warning>;
        $this->orders = <warning descr="Assign array<int,Shop\Order> to Shop\Cart::$orders, not string[].">['a']</warning>;
        self::$currency = <warning descr="Assign string to Shop\Cart::$currency, not int.">978</warning>;
        // Ranges and lengths count against exact values.
        $this->page = <warning descr="Assign positive-int to Shop\Cart::$page, not int.">0</warning>;
        $this->label = <warning descr="Assign non-empty-string to Shop\Cart::$label, not string.">''</warning>;
        // The native type rejects a value only PHPDoc describes.
        $this->total = <warning descr="Assign int to Shop\Cart::$total, not string.">$amount</warning>;
    }
}

// No template argument makes a string an array.
/** @template T */
final class Box
{
    /** @var array<int, T> */
    public $values;

    public function put() { $this->values = <warning descr="Assign array<int,mixed> to Shop\Box::$values, not string.">'x'</warning>; }
}

// A typed property declared without a default starts uninitialized, not
// null.
final class Slot
{
    /** @var Order */
    public ?object $held;

    public function free() { $this->held = <warning descr="Assign Shop\Order to Shop\Slot::$held, not null.">null</warning>; }
}
