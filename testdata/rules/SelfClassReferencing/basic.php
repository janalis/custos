<?php
namespace Shop;

use Other\Thing;

class Basket
{
    const LIMIT = Basket::class;

    public function merge(<weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning> $other, $d = <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>::LIMIT): ?<weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>
    {
        $copy = new <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">\Shop\Basket</weak_warning>();
        $copy2 = new <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">namespace\Basket</weak_warning>;
        $max  = <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>::LIMIT;
        <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>::$count++;
        <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>::create();
        if ($other instanceof <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>) {}
        log_event(<weak_warning descr="Refer to the class as &apos;__CLASS__&apos; instead of &apos;Basket::class&apos;.">Basket::class</weak_warning>);
        try {} catch (<weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning> | \Exception $e) {}

        $make = fn() => new Basket();
        $fn   = function (Basket $b): Basket { return new Basket(); };
        $anon = new class(<weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>::LIMIT) extends Basket { public function x() { return Basket::class; } };
        return self::empty();
    }

    public static function union(<weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Basket&apos;.">Basket</weak_warning>|int $x): void {}

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
    public static function make(): <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Suit&apos;.">Suit</weak_warning> { return <weak_warning descr="Refer to the class as &apos;self&apos; instead of &apos;Suit&apos;.">Suit</weak_warning>::Hearts; }
}
