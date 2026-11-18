<?php

namespace App;

use Lib\Keys;
use Lib\Mode;

function payloads(): array
{
    return [
        // Constants declared elsewhere whose value is an integer or a plain string.
        json_encode([Keys::PLAIN => Keys::QUOTED, 'n' => Keys::COUNT]),
        json_encode([\Lib\LIMIT, \Lib\LABEL]),
        // PHP's own integer constants.
        json_encode(['level' => E_ALL, 'size' => \PHP_INT_SIZE, 'flags' => \LIBXML_NOENT]),

        json_encode(Keys::RATIO, \JSON_THROW_ON_ERROR),
        json_encode(Keys::ESCAPED, \JSON_THROW_ON_ERROR),
        json_encode(Keys::JOINED, \JSON_THROW_ON_ERROR),
        json_encode(Keys::VARIABLE, \JSON_THROW_ON_ERROR),
        json_encode(Keys::MISSING, \JSON_THROW_ON_ERROR),
        json_encode(Mode::On, \JSON_THROW_ON_ERROR),
        json_encode(Missing::NAME, \JSON_THROW_ON_ERROR),
        json_encode(PHP_EOL, \JSON_THROW_ON_ERROR),
        json_encode(INF, \JSON_THROW_ON_ERROR),
    ];
}
