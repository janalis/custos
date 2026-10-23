<?php
function prune(array $groups, array $rows)
{
    // The unset after the outer loop releases the inner reference.
    foreach ($groups as $name => &$items) {
        foreach ($items as $k => &$item) {
            $item = trim($item);
        }
        if (!$items) {
            unset($groups[$name]);
        }
    }
    unset($items, $item);

    // Two levels out.
    foreach ($rows as &$row) {
        foreach ($row as &$cells) {
            foreach ($cells as &$cell) {
                $cell++;
            }
        }
    }
    unset($row, $cells, $cell);

    return [$groups, $rows];
}

function stale(array $groups)
{
    // The outer loop writes the variable before the next inner loop
    // rebinds it: the stale reference changes the last element.
    foreach ($groups as &$items) {
        $item = null;
        foreach ($items as &<warning descr="Unset '$item' right after the loop: it is still a reference to the last element.">$item</warning>) {
            $item = trim($item);
        }
        echo count($items);
    }
    unset($items, $item);

    // No unset after the outer loop either.
    foreach ($groups as $items) {
        foreach ($items as &<warning descr="Unset '$item' right after the loop: it is still a reference to the last element.">$item</warning>) {
            $item = trim($item);
        }
        echo count($items);
    }

    echo count($groups);
    return $groups;
}
