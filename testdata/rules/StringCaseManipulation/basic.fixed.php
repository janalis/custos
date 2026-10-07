<?php
function lookup(string $text, string $term, array $row) {
    $a = stripos($text, $term);
    $b = strripos($row['k'], 'abc');
    $c = mb_stripos($text, $term);
    $d = mb_strripos(trim($text), $term);
    $e = stripos($text, $term);
    return [$a, $b, $c, $d, $e];
}
