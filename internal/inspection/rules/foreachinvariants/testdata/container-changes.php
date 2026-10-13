<?php
function moveToFront(array $queue, array $vip)
{
    $len = count($queue);
    for ($i = 0; $i < $len; $i++) {
        if (in_array($queue[$i], $vip, true)) {
            array_splice($queue, $i, 1);
        }
    }
    return $queue;
}

function resorted(array $scores)
{
    $len = count($scores);
    for ($i = 0; $i < $len; $i++) {
        echo $scores[$i];
        sort($scores);
    }
}

function growing(array $jobs)
{
    for ($i = 0; $i < count($jobs); $i++) {
        if ($jobs[$i] === 'split') {
            $jobs[] = 'part';
        }
    }
    return $jobs;
}

function shrinking(array $items)
{
    for ($i = 0; $i < count($items); $i++) {
        echo $items[$i];
        unset($items[$i]);
    }
}

function replaced(array $rows)
{
    $len = count($rows);
    for ($i = 0; $i < $len; $i++) {
        echo $rows[$i];
        $rows = [];
    }
}

function elementWrites(array $cells)
{
    $len = count($cells);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $len; $i++) {
        echo $cells[$i];
    }
}
