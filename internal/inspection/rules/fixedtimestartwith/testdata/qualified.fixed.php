<?php
namespace App\Http;

function isSecure($url, $host)
{
    return [
        0 === \strncasecmp($url, 'https:', 6),
        strncmp($host, 'www.', 4) !== 0,
        Util\strpos($host, 'api.') === 0,
    ];
}
