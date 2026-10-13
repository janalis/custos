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

// mb_ conversions and searches are reported without a fix: mb_strtoupper()
// can change the number of characters ('ß' -> 'SS') and folds letters the
// case-insensitive searches do not, so matches and offsets differ.
function families(string $s, string $t) {
    $r = [];
    $r[] = <weak_warning descr="Use 'stripos($s, 'é')' instead of changing the case.">strpos(mb_strtolower($s), 'é')</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_stripos($s, $t)' instead of changing the case.">mb_strpos(strtolower($s), strtolower($t))</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_stripos($s, $t)' instead of changing the case.">mb_strpos(mb_strtoupper($s), mb_strtoupper($t))</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_strripos($s, $t)' instead of changing the case.">mb_strrpos(mb_strtolower($s), mb_strtolower($t))</weak_warning>;
    return $r;
}
