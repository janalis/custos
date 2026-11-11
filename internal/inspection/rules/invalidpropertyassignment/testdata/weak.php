<?php
// No strict_types: PHP converts scalars, but never arrays, objects or null.
namespace Shop;

class Order {}

class Account
{
    public int $balance = 0;
    public string $name = '';
    public ?Order $last = null;
}

class Savings extends Account {}

function update(Account $account, Order $order)
{
    $account->balance = '1200';
    $account->balance = 12.0;
    $account->name = 978;
    $account->balance = <error descr="Assign int to Shop\Account::$balance, not array.">[]</error>;
    $account->balance = <error descr="Assign int to Shop\Account::$balance, not null.">null</error>;
    $account->last = <error descr="Assign Shop\Order|null to Shop\Account::$last, not string.">'order-1'</error>;
    $account->last = null;
    // A subclass of Order may convert to a string.
    $account->name = $order;
}
