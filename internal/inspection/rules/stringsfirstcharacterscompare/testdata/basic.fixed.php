<?php
function route(string $uri) {
    if (strncmp($uri, '/api/', 5) === 0) {}
    if (strncmp('/admin', $uri, 6) === 0) {}
    if (\strncasecmp($uri, "HTTP\t", 5) === 0) {}
    if (strncasecmp($uri, 'C:\\x', 4) === 0) {}
    if (strncmp($uri, '/static/', 8) === 0) {}
    if (strncmp($uri, 'back\\slash', 10) === 0) {}
    if (strncmp($uri, '', 3) === 0) {}
    if (strncmp($uri, '/x', 0x2) === 0) {}
    if (strncmp($uri, '/v1/', strlen('/v1/')) === 0) {}
    if (strncmp($uri, $prefix, 4) === 0) {}
}
