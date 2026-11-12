<?php
namespace Shop;

interface Priced {}

final class Coupon
{
    public int $amount = 0;
}

class Cart implements Priced
{
    public int $total = 0;
    public static array $lines = [];

    public function sum(): int
    {
        // $this in a class.
        return $this->total + $this-><error descr="Fix the property name: Shop\Cart declares no $count.">count</error>;
    }

    public function lines(): array
    {
        // Static properties through self, the class name and static.
        $all = self::<error descr="Fix the property name: Shop\Cart declares no $rows.">$rows</error>;
        $all[] = Cart::<error descr="Fix the property name: Shop\Cart declares no $rows.">$rows</error>;
        $all[] = static::<error descr="Fix the property name: Shop\Cart declares no $rows.">$rows</error>;
        // Writing an undeclared static property fails too.
        self::<error descr="Fix the property name: Shop\Cart declares no $cache.">$cache</error> = [];
        return $all + self::$lines;
    }
}

function read(Cart $cart, ?Coupon $coupon, Priced $priced, Cart|Coupon $either): array
{
    return [
        $cart->total,
        $cart-><error descr="Fix the property name: Shop\Cart declares no $totl.">totl</error>,
        // Nullable receivers and nullsafe reads.
        $coupon-><error descr="Fix the property name: Shop\Coupon declares no $value.">value</error>,
        $coupon?-><error descr="Fix the property name: Shop\Coupon declares no $value.">value</error>,
        (new Coupon())-><error descr="Fix the property name: Shop\Coupon declares no $value.">value</error>,
        $priced-><error descr="Fix the property name: Shop\Priced declares no $price.">price</error>,
        $either-><error descr="Fix the property name: Shop\Cart|Shop\Coupon declares no $code.">code</error>,
        // Property names are case-sensitive.
        $cart-><error descr="Fix the property name: Shop\Cart declares no $Total.">Total</error>,
        // A read of an undeclared property, then of one of its elements.
        $cart-><error descr="Fix the property name: Shop\Cart declares no $items.">items</error>[0],
        $cart-><error descr="Fix the property name: Shop\Cart declares no $owner.">owner</error>->name,
        strlen($cart-><error descr="Fix the property name: Shop\Cart declares no $label.">label</error>),
        ['label' => $cart-><error descr="Fix the property name: Shop\Cart declares no $label.">label</error>],
        $cart-><error descr="Fix the property name: Shop\Cart declares no $label.">label</error> . '',
    ];
}

function guards(Coupon $coupon): ?int
{
    // A guard about another property does not cover this one.
    if (property_exists($coupon, 'amount')) {
        return $coupon-><error descr="Fix the property name: Shop\Coupon declares no $rate.">rate</error>;
    }
    return null;
}
