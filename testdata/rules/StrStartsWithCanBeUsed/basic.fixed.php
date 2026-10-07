<?php
function route(string $uri, string $prefix): array {
    return [
        str_starts_with($uri, $prefix),
        str_starts_with($uri, '/api'),
        !\str_starts_with(strtolower($uri), $prefix),
    ];
}
