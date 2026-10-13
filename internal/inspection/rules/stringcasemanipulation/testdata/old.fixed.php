<?php
// Before PHP 8.0 stripos() reads a non-string needle as a character code,
// where strtolower() made it a string: fixed only for a needle known to be
// a string.
function typed(string $text, string $term) {
    return stripos($text, $term);
}

function literal($text) {
    return strripos($text, 'abc');
}

function untyped($text, $term) {
    return strpos(strtolower($text), strtolower($term));
}

function number(string $text, int $code) {
    return strpos(strtoupper($text), strtoupper($code));
}

function haystack(string $text, $term) {
    return strpos('ABC', strtoupper($term));
}
