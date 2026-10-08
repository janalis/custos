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
    return <weak_warning descr="Replace with '$box === null'.">empty($box)</weak_warning>;
}
