<?php
namespace LocalScope;

use function extract as importRows;
use function LocalScope\extract as inspectRows;

function extract(array $rows): int { return count($rows); }

function customWriter(array $rows) {
    $count = 7;
    extract($rows);
    echo $count;
    inspectRows($rows);
    echo $count;
}

function builtinWriter(array $rows) {
    $count = 7;
    importRows($rows);
    echo (int) $count;
    $count = 11;
    echo $count;
}

function callableCreation() {
    $count = 7;
    $writer = \extract(...);
    echo $count;
    return $writer;
}
