<?php
// `$parts['port'] ?? null` is true only where $parts is an array holding
// a port.
function endpoint(string $dsn): string
{
    $parts = parse_url($dsn);
    return ($parts['port'] ?? null) ? ':' . $parts['port'] : '';
}

function scheme(string $dsn): string
{
    $parts = parse_url($dsn);
    if ($parts['scheme'] ?? false) {
        return $parts['scheme'];
    }
    return '';
}

function port(string $dsn): int
{
    $parts = parse_url($dsn);
    return ($parts['port'] ?? 0) ? $parts['port'] : 80;
}

// A fallback that holds proves nothing.
function fallback(string $dsn): string
{
    $parts = parse_url($dsn);
    return ($parts['port'] ?? 'default') ? ':' . <warning descr="Reject URL parsing failure before indexing.">$parts['port']</warning> : '';
}
