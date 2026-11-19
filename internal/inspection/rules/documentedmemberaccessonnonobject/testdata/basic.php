<?php

/**
 * @property int $count
 */
class Holder
{
    /** @var string */
    public $label;

    /** @return list<int> */
    public function ids()
    {
        return [];
    }
}

/**
 * @param int $id
 * @param string|null $name
 * @param array<string, int> $row
 * @param int|string $either
 */
function documented($id, $name, $row, $either, Holder $holder)
{
    <warning descr="Access this member on an object, not int.">$id</warning>->run();
    <warning descr="Access this member on an object, not string.">$name</warning>->value;
    <warning descr="Access this member on an object, not int[].">$row</warning>->value = 1;
    <warning descr="Access this member on an object, not int|string.">$either</warning>?->run();
    <warning descr="Access this member on an object, not string.">$holder->label</warning>->run();
    <warning descr="Access this member on an object, not int.">$holder->count</warning>->run();
    <warning descr="Access this member on an object, not int[].">$holder->ids()</warning>->run();
    /** @var float $rate */
    $rate = load();
    <warning descr="Access this member on an object, not float.">$rate</warning>->value;
}

function narrowedUnion(int|Holder $value)
{
    /** @var int $value */
    <warning descr="Access this member on an object, not int.">$value</warning>->run();
}
