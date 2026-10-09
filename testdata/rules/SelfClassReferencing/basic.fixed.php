<?php
namespace Shop;

use Other\Thing;

class Basket
{
    const LIMIT = Basket::class;

    public function merge(self $other, $d = self::LIMIT): ?self
    {
        $copy = new self();
        $copy2 = new self;
        $max  = self::LIMIT;
        self::$count++;
        Basket::create();
        if ($other instanceof self) {}
        log_event(__CLASS__);
        try {} catch (self | \Exception $e) {}

        $make = fn() => new Basket();
        $fn   = function (Basket $b): Basket { return new Basket(); };
        $anon = new class(self::LIMIT) extends Basket { public function x() { return Basket::class; } };
        return self::empty();
    }

    public static function union(self|int $x): void {}

    abstract public function nothing(Basket $b);
}

trait Countable2
{
    public function who() { return Countable2::class; }
}

interface Shape
{
    public function same(Shape $s): Shape;
}

enum Suit
{
    case Hearts;
    public static function make(): self { return self::Hearts; }
}
