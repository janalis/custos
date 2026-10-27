<?php
interface Store { public function get(string $key): mixed; }

function label(Store $m, string $to): string
{
    $to = $m->get('label') ?? $to;
    $to = str_replace("\n", '\n', $to);
    return $to;
}

function plain(string $to, $raw): string
{
    $to = $raw ?? $to;
    $to = str_replace("\n", '\n', $to);
    return $to;
}
