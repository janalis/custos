<?php
function readFirst(): string {
    $first = makeName()[0];
    return $first;
}

function wrapValue(string $value): array {
    settype($value, 'array');
    $value[] = 'extra';
    return $value;
}

function aliasedBuffer(string $buffer): array {
    $alias = &$buffer;
    $alias = [];
    $buffer[] = 'row';
    return $buffer;
}

function sharedSlot(string $slot, array &$registry) {
    $registry = ['slot' => &$slot];
    resetRegistry($registry);
    $slot[] = 'entry';
}

function resetRegistry(array &$registry) {
    $registry['slot'] = [];
}

function fillRows(&$rows, string ...$more) {
    $rows = [];
}

function collectRows(string $rows, string $label) {
    fillRows(rows: $rows);
    $rows[] = $label;
}

function collectAll(string $first, string ...$rest) {
    scanInto('a b', '%s %s', $rest, $first);
    $first[] = 'tail';
}

function scanInto(string $input, string $format, &...$targets) {
    $targets[0] = [];
}

/** @param string[] $lines */
function splitLine(array $lines) {
    settype($lines[0], 'array');
    $lines[0][] = 'part';
}
