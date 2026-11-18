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
    $plain = json_encode($q, JSON_THROW_ON_ERROR);
    $stored = [];
    $stored['x'] = json_encode($q, JSON_THROW_ON_ERROR);
    $late = json_encode($q, JSON_THROW_ON_ERROR);
    echo $late;
    $checked = false === $late;
    $over = json_encode($q, JSON_THROW_ON_ERROR);
    $over = 'x';
    $none = json_encode($q, JSON_THROW_ON_ERROR);
    // ?? does not catch json_encode()'s false.
    $coalesced = json_encode($q, JSON_THROW_ON_ERROR) ?? 'x';
    $cast = (string) json_encode($q, JSON_THROW_ON_ERROR);
    $compared = 'x' === json_encode($q, JSON_THROW_ON_ERROR);
    $other = strlen(json_encode($q, JSON_THROW_ON_ERROR));
    $this->send(json_encode($q, JSON_THROW_ON_ERROR));
    $send = 'send';
    $send(json_encode($q, JSON_THROW_ON_ERROR));
    return [$plain, $stored, $checked, $over, $coalesced, $cast, $compared, $other, fn () => $none];
}
