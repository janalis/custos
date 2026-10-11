<?php
// Before PHP 8.0 stripos() reads a non-string needle as a character code,
// where strtolower() made it a string: fixed only for a needle known to be
// a string.
function typed(string $text, string $term) {
    return <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtolower($text), strtolower($term))</weak_warning>;
}

function literal($text) {
    return <weak_warning descr="Use 'strripos($text, 'abc')' instead of changing the case.">strrpos(strtolower($text), 'abc')</weak_warning>;
}

function untyped($text, $term) {
    return <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtolower($text), strtolower($term))</weak_warning>;
}

function number(string $text, int $code) {
    return <weak_warning descr="Use 'stripos($text, $code)' instead of changing the case.">strpos(strtoupper($text), strtoupper($code))</weak_warning>;
}

function haystack(string $text, $term) {
    return <weak_warning descr="Use 'stripos('ABC', $term)' instead of changing the case.">strpos('ABC', strtoupper($term))</weak_warning>;
}
