<?php

/** @param array<int, int> $list */
function wantsList(array $list) {}

/** @param array<string, int> $map */
function wantsMap(array $map) {}

/**
 * @param array<string, int> $map
 * @param non-empty-array<string, int> $filled
 * @param non-empty-list<int> $items
 */
function keys(array $map, array $filled, array $items, $k)
{
    wantsList(<warning descr="Pass array<int,int> to $list of wantsList(), not int[].">$filled</warning>);
    wantsMap(<warning descr="Pass array<string,int> to $map of wantsMap(), not int[].">$items</warning>);
    if ($map !== []) {
        wantsList(<warning descr="Pass array<int,int> to $list of wantsList(), not int[].">$map</warning>);
    }

    wantsList($map);
    wantsMap([]);
    if (isset($map[0])) {
        wantsList($map);
    }
    if (array_key_exists(0, $map)) {
        wantsList($map);
    }
    if (array_key_exists($k, $map)) {
        wantsList($map);
    }
    if ($map[1] !== null) {
        wantsList($map);
    }
}
