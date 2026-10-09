<?php
function mutationLabels(array $values)
{
    $removed = 'before';
    unset($removed);
    $empty = '[' . (string) $removed . ']';

    $count = null;
    ++$count;
    $number = '[' . $count . ']';

    $captured = 1;
    $captured = 'after';
    $read = function () use ($captured) {
        return '[' . $captured . ']';
    };

    $dynamic = 1;
    extract($values);
    $unknown = fn() => '[' . (string) $dynamic . ']';
    return [$empty, $number, $read, $unknown];
}
