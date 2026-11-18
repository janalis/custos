<?php
// The failure value is tested where the result is produced, as an operand
// of a short-circuit operator or in a later arm of an if chain.
function shortCircuit(array $q, string $body, bool $ready): array
{
    $encoded = json_encode($q);
    $encoded || throw new RuntimeException('encoding failed');
    $decoded = json_decode($body, true);
    $decoded or throw new RuntimeException('decoding failed');
    $data = json_decode($body, true);
    if ($ready && $data) {
        return $data;
    }
    $text = json_encode($q);
    if (!$ready || !$text) {
        return [];
    }
    $rows = json_decode($body, true);
    if (!$ready) {
        return [];
    } elseif (!$rows) {
        return [];
    }
    return [$encoded, $decoded, $text, $rows];
}

// A short-circuit operand that is not a test of the result proves nothing.
function unhandled(array $q, bool $ready): array
{
    $text = json_encode($q, JSON_THROW_ON_ERROR);
    $ok = $ready && strlen($text) > 2;
    return [$ok];
}
