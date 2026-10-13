<?php
function dispatch($row, $items, $flag)
{
    $fn = 'extract';
    $fn = 'trim';
    $fn($row);

    $late = 'strtoupper';
    $late($row);
    $late = 'compact';

    $cb = 'parse_str';
    if ($flag) {
        $cb = 'trim';
    }
    array_map($cb, $items);

    foreach ($items as $step) {
        $step($row);
        $step = 'extract';
    }

    $loop = 'trim';
    foreach ($items as $item) {
        array_walk($item, $loop);
        $loop = 'trim';
    }

    foreach ($items as $item) {
        if ($item) {
            array_walk($item, <warning descr="'compact' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$next</warning>);
        }
        $next = 'compact';
    }

    $now = 'trim';
    $now = 'func_num_args';
    return <warning descr="'func_num_args' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$now</warning>();
    $now = 'extract';
}

function defaults($cb = 'extract')
{
    $cb = 'trim';
    return array_map($cb, []);
}
