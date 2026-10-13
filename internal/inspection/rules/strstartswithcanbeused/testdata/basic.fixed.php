<?php
function route(string $uri, string $prefix): array {
    return [
        str_starts_with($uri, $prefix),
        0 === mb_strpos($uri, '/api'),
        !\str_starts_with(strtolower($uri), $prefix),
    ];
}
