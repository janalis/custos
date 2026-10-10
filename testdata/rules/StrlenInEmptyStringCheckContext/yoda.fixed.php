<?php
function probe(string $name, ?string $alias, int $count, $raw, array $bag)
{
    if ('' !== $name) {}
    while (!mb_strlen($name, 'UTF-8')) {}
    if (!('' !== (string)$alias)) {}
    $a = mb_strlen($raw) ? 'y' : 'n';
    $b = $count > 2 and '' !== trim($raw);

    $c = 0 === mb_strlen($name);
    $d = '' !== (string)$count;
    $e = '' === (string)$alias;
    $f = mb_strlen($name) > 0;
    $g = '' !== (string)$bag['k'];
    $h = '' === $name;

    $i = 0 < strlen($name);
    $j = strlen($name) > 1;
    $k = (strlen($name)) > 0;
    $l = strlen($name) ?: 5;
    return strlen($name);
}

if ('' !== (string)$argv[1]) {}
