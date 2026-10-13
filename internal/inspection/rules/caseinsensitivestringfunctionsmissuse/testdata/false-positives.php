<?php
function parts($path, $mode, $flag)
{
    $two = $flag ? '/' : ':';
    $e = stripos($path, 'x');
    $f = strripos($path, 'ß');
    $g = stristr($path, '');
    $i = stripos($path, $mode);
    $j = stripos($path);
    $l = stripos($path, '/', 0, 1);
    $m = stripos($path, $two);
    $n = stripos($path, 'й');
    return [$e, $f, $g, $i, $j, $l, $m, $n];
}
$top = '/';
$o = stripos('a/b', $top);

$finder = 'stripos';
$pos = $finder($haystack, '@');
