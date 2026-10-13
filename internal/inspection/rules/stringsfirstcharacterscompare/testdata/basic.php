<?php
function route(string $uri) {
    if (strncmp($uri, '/api/', <error descr="Length 4 does not match the 5-character literal.">4</error>) === 0) {}
    if (strncmp('/admin', $uri, <error descr="Length 9 does not match the 6-character literal.">9</error>) === 0) {}
    if (\strncasecmp($uri, "HTTP\t", <error descr="Length 6 does not match the 5-character literal.">6</error>) === 0) {}
    if (strncasecmp($uri, 'C:\\x', <error descr="Length -1 does not match the 4-character literal.">-1</error>) === 0) {}
    if (strncmp($uri, '/static/', 8) === 0) {}
    if (strncmp($uri, 'back\\slash', 10) === 0) {}
    if (strncmp($uri, '', 3) === 0) {}
    if (strncmp($uri, '/x', 0x2) === 0) {}
    if (strncmp($uri, '/v1/', strlen('/v1/')) === 0) {}
    if (strncmp($uri, $prefix, 4) === 0) {}
}
