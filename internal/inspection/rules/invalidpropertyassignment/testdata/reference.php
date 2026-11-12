<?php
declare(strict_types=1);

namespace Shop;

final class Cart
{
    public int $count = 0;
    public ?int $limit = null;
    public array $items = [];

    public function bind(int $n, $unknown)
    {
        // PHP checks the local's current value when binding the reference.
        $text = 'x';
        $this->count = &<error descr="Assign int to Shop\Cart::$count, not string.">$text</error>;

        // Accepted values, or values that may be.
        $this->count = &$n;
        $this->limit = &$n;
        $this->count = &$unknown;
        // Elements and properties: a missing element is created as null.
        $this->count = &$this->items[0];
        $this->count = &$this->limit;
        // Writes through the reference after the binding.
        $number = 1;
        $this->count = &$number;
        $number = 'x';
    }

    public function bindParameter(string $s)
    {
        $this->count = &<error descr="Assign int to Shop\Cart::$count, not string.">$s</error>;
    }

    public function later()
    {
        // `$this` in a closure may be bound to another object.
        $f = function () { $text = 'x'; $this->count = &$text; };
    }
}

// A reference never initializes a readonly property.
final class Invoice
{
    public readonly int $number;

    public function __construct()
    {
        $text = 'x';
        $this->number = &$text;
    }
}
