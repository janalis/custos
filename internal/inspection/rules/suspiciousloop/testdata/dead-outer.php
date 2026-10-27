<?php
function lines(array $chunks): array
{
    $out = [];
    foreach ($chunks as $chunk) {
        foreach (explode("\n", $chunk) as $chunk) {
            $out[] = trim($chunk);
        }
    }
    return $out;
}

function files(array $dirs): iterable
{
    foreach ($dirs as $path => $depth) {
        $found = scandir($path) ?: [];
        foreach ($found as $path) {
            yield $path;
        }
    }
}

function single($rows)
{
    foreach ($rows as $row) foreach ($row as $row) echo $row;
}

function readAfter(array $rows)
{
    foreach ($rows as $row) {
        <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ($row as $row) {}
    }
    return $row;
}

function readInWrappingLoop(array $groups)
{
    while ($groups) {
        echo $row ?? '';
        foreach (array_pop($groups) as $row) {
            <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ($row as $row) {}
        }
    }
}

function notLast(array $rows)
{
    foreach ($rows as $row) {
        <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ($row as $row) {}
        echo 1;
    }
}

function byReference(array &$rows)
{
    foreach ($rows as &$row) {
        <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ([1] as $row) {}
    }
    foreach ($rows as $cell) {
        <error descr="Loop variable '$cell' overwrites a variable of an outer loop.">foreach</error> ($rows as &$cell) {}
    }
}

function emptyOuter(array $rows)
{
    foreach ($rows as $row) {}
    for ($i = 0; $i < 3; $i++) {
        <error descr="Loop variable '$i' overwrites a variable of an outer loop.">foreach</error> ($rows as $i) {}
    }
    foreach ($rows as $k) {
        <error descr="Loop variable '$k' overwrites a variable of an outer loop.">foreach</error> ($rows as $k) {}
    }
    echo $$name;
}

foreach ($top as $t) {
    foreach ($t as $t) {}
}
