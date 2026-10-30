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
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(static::class)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($this::class)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(parent::class)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Message::NAME)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['1e999'], JSON_NUMERIC_CHECK)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['a'], JSON_PRETTY_PRINT | PHP_INT_SIZE)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['a'], $rows ? 0 : JSON_PRETTY_PRINT)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['a'], 2)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode([[1]], 0, 1)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode(json_encode(['a']), true, 1)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows)</weak_warning>, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode(trim('[]'), true)</weak_warning>,
        ];
    }

    public const NAME = 'Message';
}
