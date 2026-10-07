<?php
function scan($text, $word, $flag) {
    if (strpos($text, $word) !== false) {}
    if (stripos($text, '#') === false) {}
    while ($flag and strpos($text, $word) !== false) { $flag = false; }
    $a = strpos($text, $word) !== false ? 1 : 2;
    $b = !(strpos($text, $word) !== false);
    $c = \strpos($text, $word) === false;
    $d = stripos($text, $word) !== false;
    $e = strpos($text, $word) !== false;
    do {} while ((strpos($text, $word) !== false) || $flag);
    if ($flag) {} elseif (strpos($text, $word) !== false) {}
    return [$a, $b, $c, $d, $e];
}
