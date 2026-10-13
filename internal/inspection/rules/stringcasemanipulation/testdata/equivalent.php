<?php
// Fixed only when the case-insensitive search finds exactly the same.
function find(string $text, string $term) {
    return [
        <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtolower($text), strtolower($term))</weak_warning>,
        <weak_warning descr="Use 'strripos($text, 'abc-1')' instead of changing the case.">strrpos(strtolower($text), 'abc-1')</weak_warning>,
        <weak_warning descr="Use 'stripos('ABC', $text)' instead of changing the case.">strpos('ABC', strtoupper($text))</weak_warning>,
        <weak_warning descr="Use 'stripos($text, 'Abc')' instead of changing the case.">strpos(strtolower($text), 'Abc')</weak_warning>,
        <weak_warning descr="Use 'stripos($text, 'abc')' instead of changing the case.">strpos(strtoupper($text), 'abc')</weak_warning>,
    ];
}
