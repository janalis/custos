<?php
function lookup(string $text, string $term, array $row) {
    $e = strpos(strtolower($text), $term, 3);
    $f = mb_strpos(mb_strtolower($text, 'UTF-8'), $term);
    $g = strpos(ucfirst($text), $term);
    $h = $row->strpos(strtolower($text), $term);
    $i = strpos($text, $row->strtolower($term));
    $k = strpos((strtolower($text)), $term);
    $l = strpos($text, $term);
    return [$e, $f, $g, $h, $i, $k, $l];
}
