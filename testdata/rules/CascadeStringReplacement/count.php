<?php
function normalise($text, $raw) {
    $tabs = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">["\t", "\t"]</weak_warning>, ' ', $raw, $replaced);
    $text = str_replace('a', 'b', $text);
    $text = str_replace('c', 'd', $text, $hits);
    $text = str_replace('e', 'f', $text);
    $nested = str_replace('g', 'h', str_replace('i', 'j', $raw, $inner));
    $outer = str_replace('k', 'l', str_replace('m', 'n', $raw), $total);
    return [$tabs, $text, $nested, $outer, $replaced, $hits, $inner, $total];
}
