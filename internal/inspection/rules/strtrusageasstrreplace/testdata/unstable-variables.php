<?php
function swap(string $s, bool $wide)
{
    $from = '-';
    if ($wide) {
        $from .= '-';
    }
    return strtr($s, $from, '_');
}
