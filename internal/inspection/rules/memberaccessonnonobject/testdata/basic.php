<?php

function scalars(int $id, float $rate, string $name, bool $flag, array $row, ?int $maybe, int|string $either)
{
    <error descr="Access this member on an object, not int.">$id</error>->run();
    <error descr="Access this member on an object, not float.">$rate</error>->value;
    <error descr="Access this member on an object, not string.">$name</error>->value = 1;
    <error descr="Access this member on an object, not bool.">$flag</error>?->run();
    <error descr="Access this member on an object, not array.">$row</error>->run();
    <error descr="Access this member on an object, not int.">$maybe</error>->run();
    <error descr="Access this member on an object, not int|string.">$either</error>->run();
    <error descr="Access this member on an object, not string.">'text'</error>->run();
}

function returnsTrue(): true
{
    return true;
}

function fromCall()
{
    <error descr="Access this member on an object, not true.">returnsTrue()</error>->run();
}
