<?php
// Fixed only when the case-insensitive search finds exactly the same.
function find(string $text, string $term) {
    return [
        stripos($text, $term),
        strripos($text, 'abc-1'),
        stripos('ABC', $text),
        strpos(strtolower($text), 'Abc'),
        strpos(strtoupper($text), 'abc'),
    ];
}
