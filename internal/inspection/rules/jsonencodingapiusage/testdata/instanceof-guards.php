<?php
function verified(string $payload): ?\stdClass
{
    $data = json_decode($payload);
    if (!$data instanceof \stdClass || !\is_int($data->user ?? null)) {
        return null;
    }
    return $data;
}

function signed(string|false $payload): ?\stdClass
{
    $data = false === $payload ? null : json_decode($payload);
    return $data instanceof \stdClass && isset($data->sig) ? $data : null;
}

function defaulted(string $raw): array
{
    $other = json_decode($raw, true);
    if ($other instanceof \stdClass) {
        $other = (array) $other;
    } else {
        $other = [];
    }
    return $other;
}
