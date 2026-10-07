<?php
function route(string $uri, string $prefix): array {
    return [
        <weak_warning descr="Replace with 'str_starts_with($uri, $prefix)'.">strpos($uri, $prefix) === 0</weak_warning>,
        <weak_warning descr="Replace with 'str_starts_with($uri, '/api')'.">0 === mb_strpos($uri, '/api')</weak_warning>,
        <weak_warning descr="Replace with '!\str_starts_with(strtolower($uri), $prefix)'.">\strpos(strtolower($uri), $prefix) !== 0</weak_warning>,
    ];
}
