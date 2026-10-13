<?php
// Before PHP 8.0, 0 == 'open' is true: no strict fix for numbers.
function counts(int $count, float $ratio, string $name) {
    return [
        $count == 'open',
        $ratio != 'open',
        $name === 'open',
    ];
}
