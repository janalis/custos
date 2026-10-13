<?php
function escapeAll(string $word) {
    $a = '/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">Preg_Quote</error>($word) . '/';
    $b = '/' . \<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">PREG_QUOTE</error>($word) . '/';
    $c = '/' . PREG_QUOTE($word, '/') . '/';
    return [$a, $b, $c];
}
