<?php
namespace Shop;

interface Priced
{
    public function price(): int;
}

final class Coupon {}

class Cart implements Priced
{
    public function price(): int { return 0; }
    public function add(int $n): void {}
    public static function make(): self { return new self(); }

    public function checkout(): void
    {
        // $this in a class.
        $this-><error descr="Fix the method name: Shop\Cart declares no clear().">clear</error>();
        // self::, static:: and the class name.
        self::<error descr="Fix the method name: Shop\Cart declares no build().">build</error>();
        static::<error descr="Fix the method name: Shop\Cart declares no build().">build</error>();
        Cart::<error descr="Fix the method name: Shop\Cart declares no build().">build</error>();
    }
}

final class Order extends Cart
{
    public function total(): int
    {
        return parent::<error descr="Fix the method name: Shop\Cart declares no total().">total</error>();
    }
}

function fill(Cart $cart, ?Coupon $coupon, Priced $priced, Cart|Coupon $either): void
{
    $cart->add(1);
    $cart-><error descr="Fix the method name: Shop\Cart declares no ad().">ad</error>(1);
    // Nullable receivers and nullsafe calls.
    $coupon-><error descr="Fix the method name: Shop\Coupon declares no apply().">apply</error>();
    $coupon?-><error descr="Fix the method name: Shop\Coupon declares no apply().">apply</error>();
    // A fresh object and a value a native return type gives.
    (new Coupon())-><error descr="Fix the method name: Shop\Coupon declares no apply().">apply</error>();
    Cart::make()-><error descr="Fix the method name: Shop\Cart declares no remove().">remove</error>();
    // An interface none of whose implementations declares the method.
    $priced-><error descr="Fix the method name: Shop\Priced declares no discount().">discount</error>();
    // A union none of whose members declares the method.
    $either-><error descr="Fix the method name: Shop\Cart|Shop\Coupon declares no apply().">apply</error>();
    // Method names are case-insensitive: only the missing one counts.
    $cart->ADD(2);
    // Builtin classes.
    $date = new \DateTimeImmutable();
    $date-><error descr="Fix the method name: DateTimeImmutable declares no addDays().">addDays</error>(1);
}

enum Suit: string
{
    case Hearts = 'h';
}

function cases(): void
{
    // An enum case's name is a property, not a method.
    Suit::Hearts-><error descr="Fix the method name: Shop\Suit declares no name().">name</error>();
    Suit::<error descr="Fix the method name: Shop\Suit declares no values().">values</error>();
}

function guards(Cart $cart): void
{
    // A guard about another method does not cover this one.
    if (method_exists($cart, 'remove')) {
        $cart-><error descr="Fix the method name: Shop\Cart declares no empty().">empty</error>();
    }
    // A `Class::method` string about another class or method does not either.
    if (is_callable('Shop\Coupon::empty') || is_callable('Shop\Cart::clear')) {
        $cart-><error descr="Fix the method name: Shop\Cart declares no empty().">empty</error>();
    }
}
