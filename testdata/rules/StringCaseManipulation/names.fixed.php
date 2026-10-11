<?php
namespace Search {
    function strtolower($s) { return $s; }
    function strrpos($h, $n) { return 0; }

    function find(string $text, string $term) {
        $a = strpos(strtolower($text), $term);
        $b = strrpos(\strtolower($text), $term);
        $c = stripos($text, 'term');
        return [$a, $b, $c];
    }
}

namespace {
    function find(string $text, string $term) {
        return MB_StrRPos(Mb_StrToUpper($text), MB_STRTOUPPER($term));
    }
}
