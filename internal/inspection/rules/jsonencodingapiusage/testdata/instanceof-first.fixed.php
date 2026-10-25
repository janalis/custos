<?php
function settings(?string $raw): array
{
    $values = json_decode($raw, true);
    if ($values instanceof \stdClass) {
        $values = (array) $values;
    }
    if (!\is_array($values)) {
        $values = ['personal' => []];
    }

    $other = json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
    if ($other instanceof \stdClass) {
        $other = (array) $other;
    } else {
        $other = [];
    }
    if (!\is_array($other)) {
        $other = [];
    }
    return [$values, $other];
}
