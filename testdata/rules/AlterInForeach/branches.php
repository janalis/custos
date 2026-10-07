<?php
function branches(array $items, $mode) {
    if ($mode === 1) {
        echo 'one';
    } elseif ($mode === 2) {
        foreach ($items as &$first) { $first++; }
    } else {
        echo 'other';
    }
    unset($first);

    if ($mode) foreach ($items as &$second) { $second--; } else echo 'none';
    unset($second);

    if ($mode) {
        echo 'x';
    } elseif ($items) {
        foreach ($items as &<warning descr="Unset '$third' right after the loop: it is still a reference to the last element.">$third</warning>) { $third .= '!'; }
    } else {
        echo 'y';
    }
    $third = null;

    if ($mode) {
        echo 'z';
    } else if ($items) {
        foreach ($items as &$fourth) { $fourth *= 2; }
    } else {
        echo 'w';
    }
    return $items;
}
