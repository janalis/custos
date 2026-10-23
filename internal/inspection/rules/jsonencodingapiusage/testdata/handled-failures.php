<?php
// The failure value is handled where the result is produced: throwing
// instead would change what the code does with bad input.
function handled(array $q, string $raw, string $body): array
{
    $encoded = json_encode($q);
    $key = false === $encoded ? 'none' : substr($encoded, 0, 8);
    $files = json_decode($raw, true);
    if (!\is_array($files)) {
        return [];
    }
    $a = is_array(json_decode($body, true)) ? 1 : 0;
    $b = json_decode($body, true) ?: [];
    $c = json_decode($body, true) ?? [];
    $d = (json_encode($q)) ? 'ok' : 'failed';
    if ($meta = json_decode($body, true)) {
        $a++;
    } elseif (json_decode($raw, true)) {
        $a--;
    }
    $maybe = \is_string($raw) ? json_decode($raw, true) : [];
    $rows = \is_array($maybe) ? $maybe : [];
    $copy = json_encode($q);
    while ($copy) {
        $copy = null;
    }
    return [$key, $files, $a, $b, $c, $d, $rows];
}

// Not handled: used as is, or tested only after another read, or written
// again first.
function unhandled(array $q, string $body): array
{
    $plain = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    $stored = [];
    $stored['x'] = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    $late = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    echo $late;
    $checked = false === $late;
    $over = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    $over = 'x';
    $none = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    // ?? does not catch json_encode()'s false.
    $coalesced = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning> ?? 'x';
    $cast = (string) <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    $compared = 'x' === <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>;
    $other = strlen(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
    $this->send(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
    $send = 'send';
    $send(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($q)</weak_warning>);
    return [$plain, $stored, $checked, $over, $coalesced, $cast, $compared, $other, fn () => $none];
}
