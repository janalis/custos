<?php
namespace App;

const REGION = 'fr-par';
const LIMITS = ['max' => 10, 'zone' => REGION];

// Constants whose values are literals json_encode() cannot fail on.
class Key
{
    public const ID = '1111-2222';
    private const IDS = [self::ID, 'b' => Key::ID, 'n' => -1];
    const LOOP = self::POOL;
    const POOL = self::LOOP;
    const PI = 3.14;

    public function payloads(): array
    {
        return [
            json_encode(['key_id' => self::ID, 'all' => self::IDS]),
            json_encode([REGION, LIMITS, \App\REGION]),
            json_encode(static::ID, JSON_THROW_ON_ERROR),
            json_encode(self::PI, JSON_THROW_ON_ERROR),
            json_encode(self::LOOP, JSON_THROW_ON_ERROR),
            json_encode(self::MISSING, JSON_THROW_ON_ERROR),
            json_encode(PHP_VERSION, JSON_THROW_ON_ERROR),
            json_encode(UNKNOWN_CONSTANT, JSON_THROW_ON_ERROR),
        ];
    }
}
