<?php
declare(strict_types=1);

namespace Shop;

class Order {}

final class Cart
{
    public int $count = 0;
    public string $label = '';
    /** @var int */
    public $loose = 0;
    /** @var Order|null */
    public $order = null;

    /**
     * @param array{int, string} $pair
     * @param list<string> $names
     * @param array<string, int> $byName
     */
    public function fill($pair, $names, $byName)
    {
        // Destructuring a value typed only by PHPDoc.
        [<warning descr="Assign string to Shop\Cart::$label, not int.">$this->label</warning>, <warning descr="Assign int to Shop\Cart::$count, not string.">$this->count</warning>] = $pair;
        // A list element may be missing, read as null.
        [<warning descr="Assign int to Shop\Cart::$count, not null|string.">$this->count</warning>] = $names;
        // Loops over values typed only by PHPDoc.
        foreach ($names as <warning descr="Assign int to Shop\Cart::$count, not string.">$this->count</warning>) {}
        foreach ($byName as <warning descr="Assign int to Shop\Cart::$count, not string.">$this->count</warning> => <warning descr="Assign string to Shop\Cart::$label, not int.">$this->label</warning>) {}
        // A `@var` type rejecting the stored value.
        foreach (['a'] as <warning descr="Assign int to Shop\Cart::$loose, not string.">$this->loose</warning>) {}
        [<warning descr="Assign int to Shop\Cart::$loose, not string.">$this->loose</warning>] = ['a'];
        $text = 'a';
        $this->loose = &<warning descr="Assign int to Shop\Cart::$loose, not string.">$text</warning>;
        if ($this->order === null) {
            <warning descr="Assign Shop\Order|null to Shop\Cart::$order, not int.">$this->order++</warning>;
        }

        // Accepted values.
        [$this->count, $this->label] = $pair;
        foreach ($names as $this->label) {}
        foreach ($byName as $this->label => $this->count) {}
        $this->order--;
    }
}
