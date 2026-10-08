<?php
class Box {}

function untyped() { return new Box(); }

/**
 * @param Box[] $boxes
 */
function more(array $boxes, Box $a, ?Box $b, $flag)
{
    return [
        count($boxes) === 0,
        empty(untyped()),
        ($a ?? $b) === null,
        $flag && $a === null,
        !($b !== null),
    ];
}

class Lookup
{
    /** @return Box */
    public function find($id)
    {
        return $GLOBALS['db']->get($id); // may return false despite the doc
    }
}

function docOnly(Lookup $l)
{
    $box = $l->find(1);
    return empty($box);
}
