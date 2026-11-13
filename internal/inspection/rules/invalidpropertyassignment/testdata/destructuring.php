<?php
declare(strict_types=1);

namespace Shop;

class Order {}

final class Cart
{
    public int $count = 0;
    public ?int $limit = null;
    public string $label = '';
    public ?Order $order = null;
    public static int $size = 0;

    /**
     * @param array{int, string} $documented
     */
    public function fill(array $rows, $unknown, mixed $anything, ?array $maybe, bool $flag, $documented)
    {
        // Each item stores the element of its key, positional or explicit.
        [<error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>, <error descr="Assign string to Shop\Cart::$label, not int.">$this->label</error>] = ['x', 1];
        list(<error descr="Assign string to Shop\Cart::$label, not int.">$this->label</error>, $this->count) = [1, 2];
        ['b' => <error descr="Assign string to Shop\Cart::$label, not int.">$this->label</error>, 'a' => $this->count] = ['a' => 1, 'b' => 2];
        [, <error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>] = [1, 'x'];
        // Nested patterns and static properties.
        [[<error descr="Assign int to Shop\Cart::$count, not string.">$this->count</error>], $this->label] = [['x'], 'y'];
        [<error descr="Assign int to Shop\Cart::$size, not string.">self::$size</error>] = ['x'];
        // Destructuring null or a scalar stores null in every target.
        [<error descr="Assign int to Shop\Cart::$count, not null.">$this->count</error>] = null;
        [<error descr="Assign int to Shop\Cart::$count, not null|string.">$this->count</error>] = $flag ? ['x'] : 'y';
        // A list element may be missing, read as null: both are rejected.
        $names = ['a', 'b'];
        $names[] = 'c';
        [<error descr="Assign int to Shop\Cart::$count, not null|string.">$this->count</error>] = $names;

        // Accepted values, or values that may be.
        [$this->count, $this->label] = [1, 'x'];
        [$this->limit] = null;
        [$this->limit] = $names;
        [$this->count] = $maybe;
        [$this->count] = $rows;
        [$this->count] = $unknown;
        [$this->count] = $anything;
        [$this->count] = [$unknown];
        // Only PHPDoc types the value: DocumentedPropertyAssignment's.
        [$this->label] = $documented;
        // A key the value does not list, a computed key, a reference.
        [2 => $this->count] = ['x', 'y'];
        $key = 0;
        [$key => $this->count] = ['x'];
        [&$this->count] = $rows;
        // An object may implement ArrayAccess.
        [$this->count] = new \ArrayObject(['x']);
        // Local targets are not property writes.
        [$a, $b] = ['x', 1];
    }

    public function later()
    {
        // `$this` in a closure may be bound to another object.
        $f = function () { [$this->count] = ['x']; };
    }
}

// A readonly property initialized by destructuring in its constructor.
final class Point
{
    public readonly int $x;
    public readonly int $y;

    public function __construct()
    {
        [<error descr="Assign int to Shop\Point::$x, not string.">$this->x</error>, $this->y] = ['a', 1];
    }
}

// A second write fails as a re-initialization first.
final class Segment
{
    public readonly int $from;

    public function __construct()
    {
        $this->from = 0;
        [$this->from] = ['a'];
    }
}
