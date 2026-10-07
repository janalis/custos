<?php
/** @param string[] $words */
function mangle(string $text, array $words, $other) {
    $first = $text[0][0];
    $copy  = [$text[1][0]];
    $text[2] = 'z';

    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text[3][1]</error> = 'y';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text['a']['b']</error> .= 'q';
    [<error descr="String offsets cannot be written as nested arrays (fatal error).">$text[4]['c']</error>, $tail] = $words;
    <error descr="Appending with [] is not supported on strings (fatal error).">$text[]</error> = 'w';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text[5][]</error> = 'v';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$words[0][1][2]</error> = 'u';

    <error descr="Appending with [] is not supported on strings (fatal error).">$words[0][]</error> = 't';
    $words[] = 'ok';
    $other[0][0] = 1;
    $_SESSION['cart'][] = 'item';
    $_COOKIE['a']['b'] = 'c';
}

$str = 'top';
$str[0][0] = 'x';
