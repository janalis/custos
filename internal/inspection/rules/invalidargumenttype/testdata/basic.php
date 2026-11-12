<?php
declare(strict_types=1);

namespace Shop;

final class Coupon {}
class Order {}
interface Shippable {}

function charge(int $cents, string $currency = 'EUR') {}
function ship(Shippable $item) {}
function receipt(Order $order) {}
function tally(int ...$amounts) {}

// Strict mode: scalars are not coerced.
charge(<error descr="Pass int to $cents of Shop\charge(), not string.">'1200'</error>);
charge(1200, <error descr="Pass string to $currency of Shop\charge(), not int.">978</error>);
charge(<error descr="Pass int to $cents of Shop\charge(), not float.">12.5</error>);
charge(currency: <error descr="Pass string to $currency of Shop\charge(), not true.">true</error>, cents: 5);
tally(1, 2, <error descr="Pass int to $amounts of Shop\tally(), not string.">'3'</error>);

// Classes: a final class that does not implement the interface, null.
ship(<error descr="Pass Shop\Shippable to $item of Shop\ship(), not Shop\Coupon.">new Coupon()</error>);
receipt(<error descr="Pass Shop\Order to $order of Shop\receipt(), not null.">null</error>);

// A local whose value the code shows.
function pay()
{
    $amount = 'ten';
    charge(<error descr="Pass int to $cents of Shop\charge(), not string.">$amount</error>);
}

// Methods, static methods and constructors.
class Cart
{
    public function __construct(Order $order) {}
    public function add(Order $order, int $quantity) {}
    public static function restore(array $lines) {}
}
$cart = new Cart(<error descr="Pass Shop\Order to $order of Shop\Cart::__construct(), not array.">[]</error>);
$cart->add(new Order(), <error descr="Pass int to $quantity of Shop\Cart::add(), not string.">'2'</error>);
Cart::restore(<error descr="Pass array to $lines of Shop\Cart::restore(), not string.">'a,b'</error>);

// PHP's own functions, from PHP 8.0.
strlen(<error descr="Pass string to $string of strlen(), not int.">42</error>);
str_repeat('-', <error descr="Pass int to $times of str_repeat(), not string.">'3'</error>);

// $this in a closure may be another object, but this value is a string
// whatever it is bound to.
class Ticket
{
    public function later()
    {
        return fn () => charge(<error descr="Pass int to $cents of Shop\charge(), not string.">$this instanceof Ticket ? 'a' : 'b'</error>);
    }
}
