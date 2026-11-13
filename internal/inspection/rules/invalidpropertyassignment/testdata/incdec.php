<?php
declare(strict_types=1);

namespace Shop;

class Order {}

final class Counter
{
    public null $none = null;
    public ?Order $order = null;
    public ?int $limit = null;
    public int $count = 0;
    public string $label = '';
    public bool $open = true;
    public array $items = [];

    public function bump()
    {
        // Incrementing null stores 1.
        <error descr="Assign null to Shop\Counter::$none, not int.">$this->none++</error>;
        <error descr="Assign null to Shop\Counter::$none, not int.">++$this->none</error>;
        if ($this->order === null) {
            <error descr="Assign Shop\Order|null to Shop\Counter::$order, not int.">$this->order++</error>;
        }

        // Decrementing null keeps null.
        $this->none--;
        // An integer may overflow to a float, which PHP reports itself.
        $this->count++;
        $this->limit++;
        // A numeric string becomes a number, another string its successor.
        $this->label++;
        // Booleans do not change.
        $this->open++;
        // An array or an object cannot be incremented: PHP fails before
        // storing anything.
        $this->items++;
        $this->order++;
    }

    public function later()
    {
        // `$this` in a closure may be bound to another object.
        $f = function () { $this->none++; };
    }
}

// Incrementing a readonly property reads it first: never an initialization.
// It still counts as a write, so the assignment is not the only one.
final class Ticket
{
    public readonly ?int $number;
    public readonly int $count;

    public function __construct()
    {
        $this->number++;
        $this->count = 'x';
        $this->count++;
    }
}
