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
        lookup($h->inner) === null,
        Holder::wrap($h->inner) !== null,
        empty($h->inner),
        empty($h->inner->inner),
    ];
}
