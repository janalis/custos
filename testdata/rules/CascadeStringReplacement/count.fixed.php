<?php
function normalise($text, $raw) {
    $tabs = str_replace("\t", ' ', $raw, $replaced);
    $text = str_replace('a', 'b', $text);
    $text = str_replace('c', 'd', $text, $hits);
    $text = str_replace('e', 'f', $text);
    $nested = str_replace('g', 'h', str_replace('i', 'j', $raw, $inner));
    $outer = str_replace('k', 'l', str_replace('m', 'n', $raw), $total);
    return [$tabs, $text, $nested, $outer, $replaced, $hits, $inner, $total];
}
