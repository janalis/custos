<?php
declare(strict_types=1);

namespace Shop;

final class Coupon {}
class Order {}
interface Shippable {}

// Strict mode: scalars are not coerced.
function total(): int
{
    return <error descr="Return int from Shop\total(), not string.">'1200'</error>;
}

function label(?string $fallback): string
{
    if ($fallback === null) {
        return <error descr="Return string from Shop\label(), not null.">null</error>;
    }
    return $fallback;
}

function ratio(): int
{
    return <error descr="Return int from Shop\ratio(), not float.">0.5</error>;
}

// Classes: a final class that does not implement the interface.
function parcel(): Shippable
{
    return <error descr="Return Shippable from Shop\parcel(), not Shop\Coupon.">new Coupon()</error>;
}

// A local whose value the code shows.
function amount(): int
{
    $amount = 'ten';
    return <error descr="Return int from Shop\amount(), not string.">$amount</error>;
}

// Methods, self and static.
final class Cart
{
    public function lines(): array
    {
        return <error descr="Return array from Shop\Cart::lines(), not string.">'a,b'</error>;
    }

    public function copy(): self
    {
        return <error descr="Return self from Shop\Cart::copy(), not Shop\Order.">new Order()</error>;
    }

    public function fresh(): static
    {
        return <error descr="Return static from Shop\Cart::fresh(), not Shop\Coupon.">new Coupon()</error>;
    }
}

// Closures and arrow functions answer to their own declaration.
$count = function (): int {
    return <error descr="Return int from the closure, not string.">'3'</error>;
};
$name = fn(): string => <error descr="Return string from the arrow function, not int.">42</error>;
$nested = fn(): string => (fn(): string => <error descr="Return string from the arrow function, not array.">[]</error>)();

// $this in a closure may be another object, but this value is a string
// whatever it is bound to.
class Ticket
{
    public function later()
    {
        return function (): int {
            return <error descr="Return int from the closure, not string.">$this instanceof Ticket ? 'a' : 'b'</error>;
        };
    }
}
