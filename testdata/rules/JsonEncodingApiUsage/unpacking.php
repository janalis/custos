<?php
namespace App\Unpacking;

// Unpacked arguments may supply the result type and the flags; a first-class
// callable passes none.
function unpacked(string $raw, array $all, array $rest, array $opts): array
{
    $decoder = json_decode(...);
    return [
        json_decode(...$all),
        \json_decode($raw, ...$rest),
        json_encode(['a' => 1], ...$opts),
        \json_encode(...$all),
        $decoder($raw, true, 512, JSON_THROW_ON_ERROR),
    ];
}
