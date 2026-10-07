<?php
namespace Search;

function strpos($haystack, $needle) { return 0; }
function str_contains($haystack, $needle) { return false; }

function find(string $text, string $word): array {
    $a = strpos($text, $word) !== false;
    $b = \Legacy\strpos($text, $word) !== false;
    $c = <weak_warning descr="Replace with '\str_contains($text, $word)'.">\strpos($text, $word) !== false</weak_warning>;
    $d = <weak_warning descr="Replace with '!\str_contains($text, $word)'.">mb_strpos($text, $word) === false</weak_warning>;
    return [$a, $b, $c, $d];
}
