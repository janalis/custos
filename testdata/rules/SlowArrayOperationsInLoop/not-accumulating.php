<?php
function freshPerIteration(array $rows, array $defaults)
{
    $out = [];
    foreach ($rows as $key => $row) {
        $row = array_merge($defaults, $row);
        $row['meta'] = array_merge(['kind' => 'plain'], $row['meta']);
        $out[$key] = $row;
    }
    foreach ($rows as $row) {
        $copy = $row;
        $copy = array_replace($copy, $defaults);
        $out[] = $copy;
    }
    return $out;
}

function perElement(array $tree, array $ids)
{
    foreach ($ids as $id) {
        $tree[$id] = array_merge(['open' => false], $tree[$id]);
    }
    for ($i = 0; $i < 3; $i++) {
        $tree['level'][$i] = array_merge($tree['level'][$i], ['seen' => true]);
    }
    return $tree;
}

function readBack(array $steps, array $state, array $queue, callable $expand)
{
    foreach ($steps as $step) {
        $state = array_merge($state, $step($state));
    }
    while ($next = array_pop($queue)) {
        $queue = array_merge($queue, $expand($next));
    }
    foreach ($steps as $step) {
        $state['log'] = array_merge($state['log'], [$step]);
        echo count($state['log']);
    }
    return $state;
}

class Buffer
{
    private array $items = [];

    public function fill(array $chunks): void
    {
        foreach ($chunks as $chunk) {
            $this->items = array_merge($this->items, $chunk);
            echo count($this->items);
        }
    }
}

function stillAccumulating(array $groups)
{
    $byIndex = [];
    for ($i = 0; $i < 3; $i++) {
        $byIndex = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($byIndex, $groups)</error>;
    }
    $flat = [];
    foreach ($groups as $group) {
        $flat = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($flat, $group)</error>;
    }
    $acc = ['all' => []];
    foreach ($groups as $group) {
        $acc['all'] = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($group, $acc['all'])</error>;
    }
    return [$flat, $acc, $byIndex];
}
