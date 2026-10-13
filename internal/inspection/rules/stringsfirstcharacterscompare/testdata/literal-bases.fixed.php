<?php
function route(string $uri): array {
    return [
        strncmp($uri, '/catalog', 010),
        strncmp($uri, '/api', 0x4),
        strncasecmp($uri, '/v1', 0b11),
        strncmp($uri, '/static/assets', 1_4),
        strncmp($uri, '/img', 4),
        strncmp($uri, '/docs', 5),
        strncmp($uri, '/x', 2),
    ];
}
