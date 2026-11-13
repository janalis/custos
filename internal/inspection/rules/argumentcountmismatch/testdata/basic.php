<?php
namespace Shop;

function total($items, $tax) {}
function label($text, $size = 10) {}

<error descr="Pass the required argument $tax.">total([])</error>;
<error descr="Pass the required arguments $items, $tax.">total()</error>;
<error descr="Pass the required argument $items.">total(tax: 0.2)</error>;
<error descr="Pass the required argument $text.">label(size: 12)</error>;
total([], 0.2);
label('a');

class Cart
{
    public function __construct($owner) {}
    public function add($sku, $qty) {}
    public static function restore($id) {}
    public function copy(): void
    {
        <error descr="Pass the required argument $id.">self::restore()</error>;
        <error descr="Pass the required argument $id.">static::restore()</error>;
        <error descr="Pass the required argument $qty.">$this->add('x')</error>;
    }
}

final class GiftCart extends Cart
{
    public function __construct()
    {
        <error descr="Pass the required argument $owner.">parent::__construct()</error>;
    }
}

interface Store { public function save($cart); }

function checkout(Cart $cart, Store $store, ?Cart $maybe): void
{
    <error descr="Pass the required argument $qty.">$cart->add('sku')</error>;
    <error descr="Pass the required argument $qty.">$maybe?->add('sku')</error>;
    <error descr="Pass the required argument $cart.">$store->save()</error>;
    <error descr="Pass the required argument $id.">Cart::restore()</error>;
    $cart = <error descr="Pass the required argument $owner.">new Cart</error>;
    $cart = <error descr="Pass the required argument $owner.">new Cart()</error>;
    $gift = new GiftCart();
}

<error descr="Pass the required argument $string.">strlen()</error>;
strlen('a', <error descr="Remove the extra argument: strlen() accepts at most 1.">'b'</error>);
str_repeat('-', 3, <error descr="Remove the extra arguments: str_repeat() accepts at most 2.">4, 5</error>);
$date = new \DateTimeImmutable('now', null, <error descr="Remove the extra argument: DateTimeImmutable::__construct() accepts at most 2.">'extra'</error>);
$date->format('c', <error descr="Remove the extra argument: DateTimeImmutable::format() accepts at most 1.">true</error>);
<error descr="Pass the required argument $format.">$date->format()</error>;

interface Job { public function run($at); }
class BaseJob implements Job { public function run($at) {} }
class MailJob extends BaseJob implements Job {}
function dispatch(Job $job): void
{
    <error descr="Pass the required argument $at.">$job->run()</error>;
}
