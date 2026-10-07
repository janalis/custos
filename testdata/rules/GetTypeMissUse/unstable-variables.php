<?php
function kind($value, bool $plural)
{
    $expected = 'array';
    if ($plural) {
        $expected .= 's';
    }
    $wrong = 'null';
    $wrong .= 'able';
    return [gettype($value) === $expected, gettype($value) !== $wrong];
}
