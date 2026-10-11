<?php
namespace App\ObjectAsArray;

const JSON_BIGINT_AS_STRING = 1;

function decoded(string $raw, int $mode): array
{
    return [
        // The flag already states the result type.
        json_decode($raw, flags: JSON_OBJECT_AS_ARRAY | JSON_THROW_ON_ERROR),
        json_decode($raw, flags: \JSON_THROW_ON_ERROR | (\JSON_OBJECT_AS_ARRAY)),
        json_decode($raw, flags: $mode | JSON_OBJECT_AS_ARRAY | JSON_THROW_ON_ERROR),
        // The flags may hold it: reported, not fixed.
        json_decode($raw, flags: $mode | JSON_THROW_ON_ERROR),
        json_decode($raw, flags: JSON_THROW_ON_ERROR & ~JSON_OBJECT_AS_ARRAY),
        json_decode($raw, flags: JSON_THROW_ON_ERROR | 1),
        json_decode($raw, flags: JSON_THROW_ON_ERROR | JSON_BIGINT_AS_STRING),
        json_decode($raw, flags: JSON_THROW_ON_ERROR | '0'),
        // The flags cannot hold it.
        json_decode($raw, flags: JSON_THROW_ON_ERROR, associative: false),
        json_decode($raw, flags: (JSON_THROW_ON_ERROR | \JSON_INVALID_UTF8_IGNORE) | 0, associative: false),
    ];
}
