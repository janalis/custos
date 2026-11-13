<?php
declare(strict_types=1);

namespace Shop;

class Order {}

function ship(Order $order, int $count = 1) {}
function find(): ?Order { return null; }
function label(): ?string { return null; }

class Desk
{
    private ?Order $current = null;

    public function __construct(Order $order) {}
    public function send(Order $order, string $note) {}
    public static function file(Order ...$orders) {}

    public function run(?Order $order, ?int $count): void
    {
        ship(<warning descr="Check this value for null before passing it to $order of Shop\ship(): it is Shop\Order|null.">$order</warning>);
        ship(new Order(), <warning descr="Check this value for null before passing it to $count of Shop\ship(): it is int|null.">$count</warning>);
        ship(count: <warning descr="Check this value for null before passing it to $count of Shop\ship(): it is int|null.">$count</warning>, order: new Order());
        ship(<warning descr="Check this value for null before passing it to $order of Shop\ship(): it is Shop\Order|null.">find()</warning>);
        ship(<warning descr="Check this value for null before passing it to $order of Shop\ship(): it is Shop\Order|null.">$this->current</warning>);
        $this->send(new Order(), <warning descr="Check this value for null before passing it to $note of Shop\Desk::send(): it is null|string.">label()</warning>);
        new Desk(<warning descr="Check this value for null before passing it to $order of Shop\Desk::__construct(): it is Shop\Order|null.">$order</warning>);
        Desk::file(new Order(), <warning descr="Check this value for null before passing it to $orders of Shop\Desk::file(): it is Shop\Order|null.">$order</warning>);
    }
}

// PHP's own functions reject null in strict mode.
strlen(<warning descr="Check this value for null before passing it to $string of strlen(): it is null|string.">label()</warning>);
