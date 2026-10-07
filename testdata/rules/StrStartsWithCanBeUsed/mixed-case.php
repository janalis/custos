<?php

function routed(string $path): array
{
    return [
        <weak_warning descr="Replace with 'str_starts_with($path, '/v2')'.">STRPOS($path, '/v2') === 0</weak_warning>,
        <weak_warning descr="Replace with '!str_starts_with($path, '#')'.">0 !== Mb_StrPos($path, '#')</weak_warning>,
    ];
}
