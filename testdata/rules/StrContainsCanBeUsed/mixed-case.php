<?php

function scan(string $row, string $mark): array
{
    return [
        <weak_warning descr="Replace with 'str_contains($row, $mark)'.">StrPos($row, $mark) !== false</weak_warning>,
        <weak_warning descr="Replace with '!str_contains($row, '|')'.">MB_STRPOS($row, '|') === FALSE</weak_warning>,
    ];
}
