<?php

class Holder
{
    /** @var int */
    public $total;

    /** @return string */
    public function name()
    {
        return '';
    }
}

/**
 * @param int $id
 * @param string|null $name
 * @param bool|float $either
 */
function documented($id, $name, $either, Holder $holder)
{
    foreach (<warning descr="Iterate over an array or an object, not int.">$id</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not null|string.">$name</warning> as $k => $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not bool|float.">$either</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not int.">$holder->total</warning> as $v) {
    }
    foreach (<warning descr="Iterate over an array or an object, not string.">$holder->name()</warning> as $v) {
    }
}
