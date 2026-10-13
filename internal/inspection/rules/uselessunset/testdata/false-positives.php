<?php
function consume($item, $map)
{
    foreach ($map as $entry) {
        unset($entry);
    }
    unset($map->cache, $map[0]);

    $later = function () use ($item) {
        unset($item);
    };

    if ($item) {
        return 1;
    } else {
        throw new Exception();
    }
    unset($item);
}

function early($a)
{
    return;
    unset($a);
}

unset($top);
