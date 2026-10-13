<?php

function tally(array $rows, \SplObjectStorage $seen)
{
    if (<weak_warning descr="Replace with 'count($rows) === 0'.">empty($rows)</weak_warning>) {}
    if (<weak_warning descr="Replace with 'count($seen) !== 0'.">!empty($seen)</weak_warning>) {}
    if (<weak_warning descr="Replace with 'count($rows) === 0'.">empty(($rows))</weak_warning>) {}
}

/**
 * @param float|null $ratio
 * @param true|null  $flag
 * @param string|null $label
 */
function probe($ratio, $flag, $label, ?\DateTime $when, \stdClass $box, $map)
{
    return [
        <weak_warning descr="Prefer a type-specific check over empty().">empty($ratio)</weak_warning>,
        !<weak_warning descr="Prefer a type-specific check over empty().">empty($flag)</weak_warning>,
        <weak_warning descr="Replace with '$when === null'.">empty($when)</weak_warning>,
        <weak_warning descr="Replace with '$box === null'.">empty($box)</weak_warning>,
        <weak_warning descr="Prefer a type-specific check over empty().">empty($label)</weak_warning>,
        !<weak_warning descr="Prefer a type-specific check over empty().">empty($label)</weak_warning>,
        <weak_warning descr="Prefer a type-specific check over empty().">empty(0)</weak_warning>,
        empty($map['key']),
        'x' . <weak_warning descr="Replace with '$box === null'.">empty($box)</weak_warning>,
    ];
}

class Node
{
    /** @var Node|null */
    public $next;
    /** @var string */
    public $name;
}

function walk(Node $n)
{
    return [
        empty($n->next),
        empty($n->next->next),
        <weak_warning descr="Prefer a type-specific check over empty().">empty($n->name)</weak_warning>,
    ];
}

function maybe(?int $i) { return $i; }
echo !<weak_warning descr="Prefer a type-specific check over empty().">empty(maybe(1))</weak_warning>;
