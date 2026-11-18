<?php
// Values json_encode always encodes.
function payloads(array $rows, float $f)
{
    return [
        json_encode(['key' => 'value', 'n' => -1, 2 => [true, false, null], 'nested' => ['a' => []]]),
        json_encode(("plain"), JSON_PRETTY_PRINT),
        json_encode(\NULL),
        json_encode(['iat' => time(), 'exp' => (time() + 3600) * 1, 'n' => -mt_rand(), 'd' => intdiv(7, 2) - \random_int(1, 6)]),
        json_encode(['half' => time() / 2], JSON_THROW_ON_ERROR),
        json_encode(['big' => '1e308' * '1e308'], JSON_THROW_ON_ERROR),
        json_encode(['sum' => $f + 1], JSON_THROW_ON_ERROR),
        json_encode(['pow' => time() ** 2], JSON_THROW_ON_ERROR),
        json_encode(['text' => base64_encode($f . 'x'), 'n' => \count($rows), 'sum' => md5('a')]),
        json_encode(['token' => 'T' . \count($rows), 'id' => ('a' . 1) . true]),
        json_encode(['t' => 'T' . $f], JSON_THROW_ON_ERROR),
        json_encode(['t' => 'T' . [1]], JSON_THROW_ON_ERROR),
        json_encode(['t' => 'T' + 1], JSON_THROW_ON_ERROR),
        json_encode(['raw' => md5('a', true)], JSON_THROW_ON_ERROR),
        json_encode(['k' => str_repeat('A', 1024 * 1024)]),
        json_encode(str_repeat("\xB1", 2), JSON_THROW_ON_ERROR),
        json_encode(str_repeat(['a'], 2), JSON_THROW_ON_ERROR),
        json_encode(str_repeat('a'), JSON_THROW_ON_ERROR),
        json_encode(['raw' => sha1(...['a'])], JSON_THROW_ON_ERROR),
        json_encode(['t' => trim('a')], JSON_THROW_ON_ERROR),
        json_encode(['x' => "\xB1\x31"], JSON_THROW_ON_ERROR),
        json_encode([1.5], JSON_THROW_ON_ERROR),
        json_encode(['rows' => $rows], JSON_THROW_ON_ERROR),
        json_encode([...$rows], JSON_THROW_ON_ERROR),
        json_encode([, 'a'], JSON_THROW_ON_ERROR),
        json_encode([$f => 1], JSON_THROW_ON_ERROR),
        json_encode(-$f, JSON_THROW_ON_ERROR),
        json_encode(!1, JSON_THROW_ON_ERROR),
        json_encode(PHP_EOL, JSON_THROW_ON_ERROR),
        json_encode(DEEP, JSON_THROW_ON_ERROR),
    ];
}
