<?php
namespace Search;

function strpos($haystack, $needle) { return 0; }
function str_contains($haystack, $needle) { return false; }

function find(string $text, string $word): array {
    $a = strpos($text, $word) !== false;
    $b = \Legacy\strpos($text, $word) !== false;
    $c = \str_contains($text, $word);
    $d = !\str_contains($text, $word);
    return [$a, $b, $c, $d];
}
