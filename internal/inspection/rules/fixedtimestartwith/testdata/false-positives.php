<?php
function prefixes($url, $part, $o)
{
    return [
        strpos($url, 'https') == 0,
        strpos($url, 'https', 2) === 0,
        strpos($url, $part) === 0,
        strpos($url, "{$part}/") === 0,
        (strpos($url, 'https')) === 0,
        strpos($url, 'https') === false,
        strpos($url, 'https') === 00,
        strpos($url, 'https') === 0.0,
        $o->strpos($url, 'https') === 0,
        strpos($url, 'a' . 'b') === 0,
    ];
}
