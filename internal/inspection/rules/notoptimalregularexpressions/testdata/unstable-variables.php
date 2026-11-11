<?php
function patterns(string $s, bool $ci)
{
    $re = '/[0-9]+/';
    if ($ci) {
        $re .= 'i';
    }
    $plain = 'abc';
    $plain .= 'def';
    // Two possible patterns: neither is the one checked.
    $either = $ci ? '/[0-9]+/' : '/[0-9]+/i';
    return [preg_match($re, $s), preg_match('#' . $plain . '#', $s), preg_split([$re], $s), preg_match($either, $s)];
}
