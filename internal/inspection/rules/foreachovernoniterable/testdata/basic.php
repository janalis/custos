<?php

function scalars(int $id, float $rate, string $name, bool $flag, ?int $maybe, null $nothing, int|string $either)
{
    foreach (<warning descr="Iterate over an array or an object, not int.">$id</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not float.">$rate</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not string.">$name</warning> as $k => $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not bool.">$flag</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not int|null.">$maybe</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not null.">$nothing</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not int|string.">$either</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not int.">3</warning> as $v) {
    }
}

function returnsFalse(): false
{
    return false;
}

function fromCall()
{
    foreach (<warning descr="Iterate over an array or an object, not false.">returnsFalse()</warning> as $v) {
    }
}

function returnsTrue(): true
{
    return true;
}

function fromOtherCall()
{
    foreach (<warning descr="Iterate over an array or an object, not true.">returnsTrue()</warning> as $v) {
    }
}
