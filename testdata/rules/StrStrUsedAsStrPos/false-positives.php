<?php
function scan($text, $word, $o) {
    $e = strstr($text, $word);
    $f = true == strstr($text, $word);
    $g = (strstr($text, $word)) === false;
    $h = strstr($text, $word) ?: 'none';
    $i = strstr($text);
    $k = $o->strstr($text, $word) ? 1 : 0;
    $l = $e xor strstr($text, $word);
    $m = strstr($text, $word) === '';
    $n = strstr($text, $word) < false;
    for (; strstr($text, $word);) {}
    $p = $o ? strstr($text, $word) : null;
    $q = strstr(needle: $word, haystack: $text) ? 1 : 0;
    $r = strstr($text, ...[$word]) ? 1 : 0;
    $s = strstr($text, $word, true) ? 'prefix' : 'none';
    $t = !strstr($text, $word, true);
    $u = strstr($text, $word, true) == false;
    $v = strstr($text, $word, $o) === false;
    $w = strstr($text, $word, false, 1) ? 1 : 0;
    return strstr($text, $word);
}
