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

// Conversions of the other family: mb_strtolower() folds non-ASCII letters,
// stripos() does not (and mb_stripos() does where strtolower() does not).
function families(string $s, string $t) {
    $r = [];
    $r[] = <weak_warning descr="Use 'stripos($s, 'é')' instead of changing the case.">strpos(mb_strtolower($s), 'é')</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_stripos($s, $t)' instead of changing the case.">mb_strpos(strtolower($s), strtolower($t))</weak_warning>;
    return $r;
}
