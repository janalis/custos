<?php

function tally(array $rows, \SplObjectStorage $seen)
{
    if (count($rows) === 0) {}
    if (count($seen) !== 0) {}
    if (count($rows) === 0) {}
}

/**
 * @param float|null $ratio
 * @param true|null  $flag
 * @param string|null $label
 */
function probe($ratio, $flag, $label, ?\DateTime $when, \stdClass $box, $map)
{
    return [
        empty($ratio),
        !empty($flag),
        $when === null,
        $box === null,
        empty($label),
        !empty($label),
        empty(0),
        empty($map['key']),
        'x' . ($box === null),
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
        empty($n->name),
    ];
}

function maybe(?int $i) { return $i; }
echo !empty(maybe(1));
