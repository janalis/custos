<?php
// null, false and true count as 0, 0 and 1 in arithmetic: an int result.
function deadline(string $start, ?int $offset, string $label) {
    return [
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> (strtotime($start) + 60),
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> ($offset * 2),
        (int) (strtotime($start) / 60),
        (int) ($label + 1),
    ];
}
