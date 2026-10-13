<?php
function lookup(string $text, string $term, array $row) {
    $a = strpos(strtoupper($text), $term);
    $b = strrpos($row['k'], strtolower('abc'));
    $c = \mb_strpos(mb_strtolower($text), mb_strtoupper($term));
    $d = mb_strrpos(strtoupper(trim($text)), mb_strtolower($term));
    $e = strpos(strtolower($text),$term);
    return [$a, $b, $c, $d, $e];
}
