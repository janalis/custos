<?php
function split_parts(string $line): array
{
    $a = strpos($line, "\t");
    $b = strrpos($line, "\x2F");
    $c = strstr($line, "\u{2014}\n");
    $d = strpos($line, '\\');
    $e = stripos($line, "\x41");
    $f = stripos($line, "\u{e9}");
    $g = stripos($line, "\xE9");
    $h = stripos($line, '\t');
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}
