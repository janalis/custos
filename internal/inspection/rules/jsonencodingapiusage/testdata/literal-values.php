<?php
// Values json_encode always encodes.
function payloads(array $rows, float $f)
{
    return [
        json_encode(['key' => 'value', 'n' => -1, 2 => [true, false, null], 'nested' => ['a' => []]]),
        json_encode(("plain"), JSON_PRETTY_PRINT),
        json_encode(\NULL),
        json_encode(['iat' => time(), 'exp' => (time() + 3600) * 1, 'n' => -mt_rand(), 'd' => intdiv(7, 2) - \random_int(1, 6)]),
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['half' => time() / 2])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['big' => '1e308' * '1e308'])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['sum' => $f + 1])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['pow' => time() ** 2])</weak_warning>,
        json_encode(['text' => base64_encode($f . 'x'), 'n' => \count($rows), 'sum' => md5('a')]),
        json_encode(['token' => 'T' . \count($rows), 'id' => ('a' . 1) . true]),
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['t' => 'T' . $f])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['t' => 'T' . [1]])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['t' => 'T' + 1])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['raw' => md5('a', true)])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['raw' => sha1(...['a'])])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['t' => trim('a')])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['x' => "\xB1\x31"])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode([1.5])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['rows' => $rows])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode([...$rows])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode([, 'a'])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode([$f => 1])</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(-$f)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(!1)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(PHP_EOL)</weak_warning>,
        <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(DEEP)</weak_warning>,
    ];
}
