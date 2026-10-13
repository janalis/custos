<?php
namespace Search {
    function stripos($haystack, $needle) { return 0; }

    function find(string $text, string $term) {
        return \stripos($text, 'term');
    }
}
