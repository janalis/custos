<?php
// Division of two ints returns a float unless it is exact: the cast matters.
function shares(int $total, int $parts, float $weight) {
    $each = $total / $parts;
    return [
        (int) ($total / $parts),
        (int) ($total / 2 + 1),
        (int) $each,
        (float) ($total / $parts),
        ($weight / $parts),
        ($total * $parts),
    ];
}
