<?php
function parse(array $tags): array
{
    foreach ($tags['method'] ?? [] as $i => $method) {
        unset($tags['method'][$i]);
        for ($i = 0; $i < 3; $i++) {
            echo $method, $i;
        }
    }
    foreach ($tags['param'] ?? [] as $i => $param) {
        echo $i, $param;
    }
    return $tags;
}
function render(\Throwable $e): void
{
    do {
        $lines = [];
        foreach (explode("\n", $e->getMessage()) as $line) {
            foreach (str_split($line, 10) as $line) {
                $lines[] = $line;
            }
        }
        foreach ($lines as $line) {
            echo $line;
        }
    } while ($e = $e->getPrevious());
}

function subjectRead(array $rows)
{
    foreach ($rows as $row) {
        <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ($row as $row) {}
    }
    foreach ($row as $row) {}
}

function byRefLater(array $rows)
{
    foreach ($rows as $row) {
        <error descr="Loop variable '$row' overwrites a variable of an outer loop.">foreach</error> ($row as $row) {}
    }
    foreach ($rows as &$row) {
        echo $row;
    }
}

function counterLater(array $rows, int $i)
{
    echo $i;
    foreach ($rows as $i) {}
    for ($i = 0; $i < 3; $i++) {
        echo $i;
    }
}
