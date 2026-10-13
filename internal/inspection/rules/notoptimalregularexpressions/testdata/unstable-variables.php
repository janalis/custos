<?php
function patterns(string $s, bool $ci)
{
    $re = '/[0-9]+/';
    if ($ci) {
        $re .= 'i';
    }
    $plain = 'abc';
    $plain .= 'def';
    return [preg_match($re, $s), preg_match('#' . $plain . '#', $s), preg_split([$re], $s)];
}
