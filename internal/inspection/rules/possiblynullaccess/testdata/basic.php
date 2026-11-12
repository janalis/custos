<?php
namespace Shop;

class Item
{
    public const KIND = 'item';
    public static int $made = 0;
    public int $qty = 0;

    public function price(): int { return 1; }
    public static function make(): self { return new self(); }
}

class Cart
{
    private ?Item $last = null;

    public function find(): ?Item { return null; }

    // A nullable parameter: every kind of access.
    public function each(?Item $item, ?array $lines): void
    {
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>->qty = 2;
        echo <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>->qty;
        echo <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>::KIND;
        echo <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>::$made;
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>::make();
        echo <warning descr="Check this value for null before the access: it is array|null.">$lines</warning>['a'];
    }

    // A nullable property, return and local.
    public function sources(): void
    {
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->last</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->find()</warning>->price();
        $found = $this->find();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$found</warning>->price();
    }

    // Only null.
    public function never(): void
    {
        $item = null;
        <warning descr="Check this value for null before the access: it is null.">$item</warning>->price();
    }

    // A check that does not hold on every path.
    public function partial(?Item $item, bool $log): void
    {
        if ($item === null) {
            echo 'missing';
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>->price();
    }

    // A null assigned on one branch.
    public function branch(bool $ok): void
    {
        $item = new Item();
        if (!$ok) {
            $item = null;
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item</warning>->price();
    }

    // A property read on an element of a nullable array is still a read.
    public function nested(?array $rows): void
    {
        echo <warning descr="Check this value for null before the access: it is array|null.">$rows</warning>['a']['b'];
        $first = <warning descr="Check this value for null before the access: it is array|null.">$rows</warning>[0] + 1;
    }
}
