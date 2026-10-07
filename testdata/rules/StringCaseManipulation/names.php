<?php
namespace Search {
    function strtolower($s) { return $s; }
    function strrpos($h, $n) { return 0; }

    function find(string $text, string $term) {
        $a = strpos(strtolower($text), $term);
        $b = strrpos(\strtolower($text), $term);
        $c = <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">\STRPOS(\StrToLower($text), $term)</weak_warning>;
        return [$a, $b, $c];
    }
}

namespace {
    function find(string $text, string $term) {
        return <weak_warning descr="Use 'mb_strripos($text, $term)' instead of changing the case.">MB_StrRPos(Mb_StrToUpper($text), $term)</weak_warning>;
    }
}
