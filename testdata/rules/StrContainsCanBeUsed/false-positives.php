<?php
function scan(string $line, string $tag) {
    $from = strpos($line, $tag, 3) !== false;
    $loose = strpos($line, $tag) != false;
    $zero = strpos($line, $tag) === 0;
    $null = strpos($line, $tag) !== null;
    $wrapped = (strpos($line, $tag)) !== false;
    $other = stripos($line, $tag) !== false;
    $enc = mb_strpos($line, $tag, 0, 'UTF-8') !== false;
    return [$from, $loose, $zero, $null, $wrapped, $other, $enc];
}
