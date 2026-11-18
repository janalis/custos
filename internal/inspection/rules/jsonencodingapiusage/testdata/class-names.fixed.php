<?php
namespace App;

// Class names, and text a json_encode() that cannot fail produced.
final class Message
{
    public function payloads(array $rows): array
    {
        return [
            json_encode(['type' => Message::class, 'self' => self::class, 'other' => \stdClass::class]),
            json_encode(['type' => Message::class], JSON_PRETTY_PRINT | (JSON_UNESCAPED_SLASHES | 0)),
            json_decode(json_encode(['data' => 'https://host/page?id=5']), false),
            json_decode((json_encode([1, [2]], JSON_PRETTY_PRINT)), true, flags: JSON_BIGINT_AS_STRING),
            json_encode(static::class, JSON_THROW_ON_ERROR),
            json_encode($this::class, JSON_THROW_ON_ERROR),
            json_encode(parent::class, JSON_THROW_ON_ERROR),
            json_encode(['1e999'], JSON_THROW_ON_ERROR | JSON_NUMERIC_CHECK),
            json_encode(['a'], JSON_THROW_ON_ERROR | JSON_PRETTY_PRINT | PHP_INT_SIZE),
            json_encode(['a'], JSON_THROW_ON_ERROR | ($rows ? 0 : JSON_PRETTY_PRINT)),
            json_encode(['a'], JSON_THROW_ON_ERROR | 2),
            json_encode([[1]], JSON_THROW_ON_ERROR | 0, 1),
            json_decode(json_encode(['a']), true, 1, JSON_THROW_ON_ERROR),
            json_decode(json_encode($rows), true, 512, JSON_THROW_ON_ERROR),
            json_decode(trim('[]'), true, 512, JSON_THROW_ON_ERROR),
        ];
    }

    public const NAME = 'Message';
}
