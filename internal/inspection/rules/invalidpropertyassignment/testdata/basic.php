<?php
declare(strict_types=1);

namespace Shop;

final class Coupon {}
class Order {}
interface Shippable {}

final class Cart
{
    public int $count = 0;
    public ?Order $order = null;
    public Shippable $parcel;
    public static array $lines = [];
    protected string $label = '';
    private float $total = 0.0;

    public function fill(Coupon $coupon)
    {
        // Strict mode: scalars are not converted.
        $this->count = <error descr="Assign int to Shop\Cart::$count, not string.">'3'</error>;
        $this->label = <error descr="Assign string to Shop\Cart::$label, not int.">42</error>;
        $this->total = <error descr="Assign float to Shop\Cart::$total, not string.">'9.99'</error>;
        // Classes: a final class that does not implement the interface.
        $this->order = <error descr="Assign Shop\Order|null to Shop\Cart::$order, not Shop\Coupon.">$coupon</error>;
        $this->parcel = <error descr="Assign Shop\Shippable to Shop\Cart::$parcel, not Shop\Coupon.">$coupon</error>;
        // Static properties through self, static and the class name.
        self::$lines = <error descr="Assign array to Shop\Cart::$lines, not string.">'a,b'</error>;
        static::$lines = <error descr="Assign array to Shop\Cart::$lines, not null.">null</error>;
        Cart::$lines = <error descr="Assign array to Shop\Cart::$lines, not int.">1</error>;
        // Compound assignments store the result of the operation.
        $this->count .= <error descr="Assign int to Shop\Cart::$count, not string.">'1'</error>;
        // `??=` stores its right-hand side.
        $this->order ??= <error descr="Assign Shop\Order|null to Shop\Cart::$order, not string.">'none'</error>;
    }
}

// Through a receiver whose type the code declares.
function reset(Cart $cart)
{
    $cart->count = <error descr="Assign int to Shop\Cart::$count, not null.">null</error>;
    $amount = 'ten';
    $cart->count = <error descr="Assign int to Shop\Cart::$count, not string.">$amount</error>;
    // A closure may run in another scope; a public property stays writable.
    $later = function () use ($cart) { $cart->count = <error descr="Assign int to Shop\Cart::$count, not string.">'x'</error>; };
}

// Inherited and trait-declared properties.
trait Tagged
{
    public string $tag = '';
}

class Base
{
    use Tagged;

    public int $level = 0;
}

final class Leaf extends Base
{
    public function bump()
    {
        $this->level = <error descr="Assign int to Shop\Base::$level, not array.">[]</error>;
        $this->tag = <error descr="Assign string to Shop\Base::$tag, not int.">1</error>;
    }
}

// A readonly property initialized by its own constructor.
final class Invoice
{
    public readonly int $number;

    public function __construct()
    {
        $this->number = <error descr="Assign int to Shop\Invoice::$number, not string.">'A-1'</error>;
    }
}

// `static::` may name a subclass, which inherits the declaration.
class Registry
{
    public static int $size = 0;

    public static function grow() { static::$size = <error descr="Assign int to Shop\Registry::$size, not string.">'x'</error>; }
}

class Narrow extends Registry {}
