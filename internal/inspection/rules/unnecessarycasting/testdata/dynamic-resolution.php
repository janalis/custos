<?php
namespace LocalScope;

use function extract as importRows;
use function LocalScope\extract as inspectRows;

function extract(array $rows): int { return count($rows); }

function customWriter(array $rows) {
    $count = 7;
    extract($rows);
    echo <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $count;
    inspectRows($rows);
    echo <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $count;
}

function builtinWriter(array $rows) {
    $count = 7;
    importRows($rows);
    echo (int) $count;
    $count = 11;
    echo <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $count;
}

function callableCreation() {
    $count = 7;
    $writer = \extract(...);
    echo <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $count;
    return $writer;
}
