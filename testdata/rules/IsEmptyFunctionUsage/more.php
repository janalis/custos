<?php
class Box {}

function untyped() { return new Box(); }

/**
 * @param Box[] $boxes
 */
function more(array $boxes, Box $a, ?Box $b, $flag)
{
    return [
        <weak_warning descr="Replace with 'count($boxes) === 0'.">empty($boxes)</weak_warning>,
        <weak_warning descr="Replace with 'untyped() === null'.">empty(untyped())</weak_warning>,
        <weak_warning descr="Replace with '($a ?? $b) === null'.">empty($a ?? $b)</weak_warning>,
        $flag && <weak_warning descr="Replace with '$a === null'.">empty($a)</weak_warning>,
        !<weak_warning descr="Replace with '$b !== null'.">!empty($b)</weak_warning>,
    ];
}
