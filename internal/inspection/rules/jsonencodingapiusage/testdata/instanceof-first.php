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

    $other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>;
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
