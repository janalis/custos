<?php declare(strict_types=1);
// Under strict types a null given for a string parameter throws a TypeError.
function clean(string $text): string
{
    $text = preg_replace('/\s+/', ' ', $text);
    return trim($text);
}
