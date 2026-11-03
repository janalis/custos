<?php
function last(array $items, object $stub)
{
    foreach ($items as $stub->class => $stub->position) {}
    foreach ($items as [$a, $b]) {}
    foreach ($items as $key => $item) {}
    return $key;
}

// By reference, the loop leaves every entry of the array a reference.
function referenced(array $keys, $redis)
{
    foreach ($keys as &$key) {}
    return $redis->hMget('h', $keys);
}

function referencedUnused(array $keys, array $other, array $more)
{
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($keys as &$key) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($more as $plain) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($other['k'] as &$entry) {}
    return [$more, $other];
}

function unused(array $items)
{
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($items as $key => $item) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($items as $other) {}
    echo $items;
}

foreach ($top as $t) {}
echo $t;
