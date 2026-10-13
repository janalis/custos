<?php
function route(string $uri, string $prefix): array {
    return [
        strpos($uri, $prefix, 1) === 0,
        strpos($uri, $prefix) == 0,
        strpos($uri, $prefix) === 00,
        strpos($uri, $prefix) === 0.0,
        strpos($uri, $prefix) === '0',
        strpos($uri, $prefix) === false,
        (strpos($uri, $prefix)) === 0,
        stripos($uri, $prefix) === 0,
    ];
}
