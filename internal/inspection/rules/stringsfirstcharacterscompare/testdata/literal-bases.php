<?php
function route(string $uri): array {
    return [
        strncmp($uri, '/catalog', 010),
        strncmp($uri, '/api', 0x4),
        strncasecmp($uri, '/v1', 0b11),
        strncmp($uri, '/static/assets', 1_4),
        strncmp($uri, '/img', <error descr="Length 010 does not match the 4-character literal.">010</error>),
        strncmp($uri, '/docs', <error descr="Length 0x10 does not match the 5-character literal.">0x10</error>),
        strncmp($uri, '/x', <error descr="Length 1_0 does not match the 2-character literal.">1_0</error>),
    ];
}
