<?php
namespace Search {
    function stripos($haystack, $needle) { return 0; }

    function find(string $text, string $term) {
        return <weak_warning descr="Use '\stripos($text, 'term')' instead of changing the case.">strpos(strtolower($text), 'term')</weak_warning>;
    }
}
