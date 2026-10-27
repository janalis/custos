<?php
function last(array $items, object $stub)
{
    foreach ($items as $stub->class => $stub->position) {}
    foreach ($items as [$a, $b]) {}
    foreach ($items as $key => $item) {}
    return $key;
}

function unused(array $items)
{
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($items as $key => $item) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($items as $other) {}
    echo $items;
}

foreach ($top as $t) {}
echo $t;
