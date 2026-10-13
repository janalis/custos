<?php
namespace App\Http;

function isSecure($url, $host)
{
    return [
        0 === <warning descr="Use '\strncasecmp($url, 'https:', 6)' for a length-independent prefix check.">\stripos($url, 'https:')</warning>,
        <warning descr="Use 'strncmp($host, 'www.', 4)' for a length-independent prefix check.">strpos($host, 'www.')</warning> !== 0,
        Util\strpos($host, 'api.') === 0,
    ];
}
