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
        untyped() === null,
        ($a ?? $b) === null,
        $flag && $a === null,
        !($b !== null),
    ];
}
