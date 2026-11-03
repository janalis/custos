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

        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Keys::RATIO)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Keys::ESCAPED)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Keys::JOINED)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Keys::VARIABLE)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Keys::MISSING)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Mode::On)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Missing::NAME)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(PHP_EOL)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(INF)</weak_warning>,
    ];
}
