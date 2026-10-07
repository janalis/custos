<?php
function collect($flag, $items, $row)
{
    $either = $flag ? 'extract' : 'compact';
    $either($row);

    if ($flag) {
        $twice = 'extract';
    } else {
        $twice = 'extract';
    }
    $twice($row);

    array_filter($items);
    array_filter($items, 'trim');
    array_map('parse_str ', $items);
    array_map(function ($x) { return $x; }, $items);
    compact('either', 'twice');
    \extract($row);
}

$top = 'get_defined_vars';
$top();
