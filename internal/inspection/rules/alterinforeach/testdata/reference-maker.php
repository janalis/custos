<?php
function bind(array $keys): array
{
    // turns every element into a reference
    foreach ($keys as &$key) {}
    consume($keys);
    return $keys;
}

function reused(array $keys): void
{
    foreach ($keys as &<warning descr="Unset '$key' right after the loop: it is still a reference to the last element.">$key</warning>) {}
    $key = 'x';
}

function looped(array $groups): void
{
    while ($groups) {
        $key = array_pop($groups);
        foreach ($groups as &<warning descr="Unset '$key' right after the loop: it is still a reference to the last element.">$key</warning>) {}
        consume($groups);
    }
}

function dynamic(array $keys, string $n): void
{
    foreach ($keys as &<warning descr="Unset '$key' right after the loop: it is still a reference to the last element.">$key</warning>) {}
    $$n = 1;
}

function nonEmpty(array $keys): void
{
    foreach ($keys as &<warning descr="Unset '$key' right after the loop: it is still a reference to the last element.">$key</warning>) { $key++; }
    consume($keys);
}

foreach ($top as &$t) {}
consume($top);
