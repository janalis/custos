<?php
function lookup(string $text, string $term, array $row) {
    $a = <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtoupper($text), $term)</weak_warning>;
    $b = <weak_warning descr="Use 'strripos($row['k'], 'abc')' instead of changing the case.">strrpos($row['k'], strtolower('abc'))</weak_warning>;
    $c = <weak_warning descr="Use 'mb_stripos($text, $term)' instead of changing the case.">\mb_strpos(mb_strtolower($text), mb_strtoupper($term))</weak_warning>;
    $d = <weak_warning descr="Use 'mb_strripos(trim($text), $term)' instead of changing the case.">mb_strrpos(strtoupper(trim($text)), mb_strtolower($term))</weak_warning>;
    $e = <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtolower($text),$term)</weak_warning>;
    return [$a, $b, $c, $d, $e];
}
