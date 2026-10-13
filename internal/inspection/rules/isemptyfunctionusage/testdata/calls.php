<?php
class Holder
{
    /** @var Holder|null */
    public $inner;

    public static function wrap($v): ?Holder { return $v; }
}

function lookup($key): ?\DateTime { return null; }

function scan(Holder $h)
{
    return [
        <weak_warning descr="Replace with 'lookup($h->inner) === null'.">empty(lookup($h->inner))</weak_warning>,
        <weak_warning descr="Replace with 'Holder::wrap($h->inner) !== null'.">!empty(Holder::wrap($h->inner))</weak_warning>,
        empty($h->inner),
        empty($h->inner->inner),
    ];
}
