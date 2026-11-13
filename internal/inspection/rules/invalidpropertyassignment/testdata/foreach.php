<?php
declare(strict_types=1);

namespace Shop;

final class Cart
{
    public int $count = 0;
    public string $label = '';
    public bool $open = true;
    public int|string $key = 0;
    public array $items = [];

    /**
     * @param list<string> $documented
     */
    public function fill(array $rows, $unknown, \Traversable $lines, \Generator $pages, $documented)
    {
        // Each iteration stores the element, and the key.
        foreach (['a', 'b'] as <error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>) {}
        foreach (['a' => 1] as <error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error> => <error descr="Assign string to Shop\Cart::$label, not int.">$this->label</error>) {}
        $codes = ['x', 'y'];
        foreach ($codes as $code => <error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>) {}
        // Array keys are integers or strings.
        foreach ($rows as <error descr="Assign bool to Shop\Cart::$open, not int|string.">$this->open</error> => $row) {}
        // A destructured element.
        foreach ([['x', 1]] as [<error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>, <error descr="Assign string to Shop\Cart::$label, not int.">$this->label</error>]) {}

        // Accepted values, or values that may be.
        foreach (['a' => 1] as $this->label => $this->count) {}
        foreach ($rows as $this->key => $this->count) {}
        foreach ($unknown as $this->open => $this->count) {}
        // An iterator may yield keys and values of any type.
        foreach ($lines as $this->open => $this->count) {}
        foreach ($pages as $this->open => $this->count) {}
        // Only PHPDoc types the elements: DocumentedPropertyAssignment's.
        foreach ($documented as $this->count) {}
        // By-reference values bind a reference to each element.
        $more = ['x'];
        foreach ($more as &$this->count) {}
        // Element writes, not property writes.
        foreach ($codes as $this->items[]) {}
    }

    public function later()
    {
        // `$this` in a closure may be bound to another object.
        $f = function () { foreach (['a'] as $this->count) {} };
        $g = function (array $rows) { foreach ($rows as $this->open => $row) {} };
        $h = function () { foreach ($this->items as $this->count) {} };
    }
}

// A readonly property bound by a loop in its constructor: the first
// iteration initializes it.
final class Cursor
{
    public readonly int $position;
    public readonly int $index;

    public function __construct()
    {
        foreach (['a'] as <error descr="Assign int to Shop\Cursor::$position, not string.">$this->position</error>) {}
        foreach (['a' => 1] as $this->index => $value) {}
        $this->index = 0;
    }
}
