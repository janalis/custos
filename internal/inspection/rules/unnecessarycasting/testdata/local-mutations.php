<?php
function mutationLabels(array $values)
{
    $removed = 'before';
    unset($removed);
    $empty = '[' . (string) $removed . ']';

    $count = null;
    ++$count;
    $number = '[' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $count . ']';

    $captured = 1;
    $captured = 'after';
    $read = function () use ($captured) {
        return '[' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $captured . ']';
    };

    $dynamic = 1;
    extract($values);
    $unknown = fn() => '[' . (string) $dynamic . ']';
    return [$empty, $number, $read, $unknown];
}
