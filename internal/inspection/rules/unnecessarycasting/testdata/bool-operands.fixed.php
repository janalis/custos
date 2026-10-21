<?php
// null, false and true count as 0, 0 and 1 in arithmetic: an int result.
function deadline(string $start, ?int $offset, string $label) {
    return [
        (strtotime($start) + 60),
        ($offset * 2),
        (int) (strtotime($start) / 60),
        (int) ($label + 1),
    ];
}
