<?php
namespace App\Settings;

// The file writes global constants qualified (\PHP_EOL): so does the fix.
function rows(string $raw, array $v, bool $pretty, int $a, int $b): array
{
    return [
        \json_decode($raw, associative: true, flags: \JSON_THROW_ON_ERROR),
        json_encode($v, \JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES),
        \json_encode($v, flags: \JSON_THROW_ON_ERROR | ($pretty ? \JSON_PRETTY_PRINT : 0)),
        \json_encode($v, \JSON_THROW_ON_ERROR | ($pretty ? \JSON_PRETTY_PRINT : 0)),
        \json_encode($v, \JSON_THROW_ON_ERROR | $a | $b),
        \json_encode(value: $v, flags: \JSON_THROW_ON_ERROR),
        \json_encode($v, \JSON_THROW_ON_ERROR | ($a ?: $b)),
        \json_encode($v, \JSON_THROW_ON_ERROR | ($a ?? $b)),
        \json_decode($raw, depth: 8, associative: true, flags: \JSON_THROW_ON_ERROR),
        \json_decode(
            $raw,
            associative: false, flags: \JSON_THROW_ON_ERROR,
        ),
        \PHP_EOL,
    ];
}
