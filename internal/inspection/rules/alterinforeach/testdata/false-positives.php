<?php
function emptyStatements(array $xs) {
    foreach ($xs as &$v) { $v = trim($v); };
    unset($v);
    foreach ($xs as &$w) { $w++; } ; ;
    return $xs;
}

function leavesSwitch(array $xs, $mode) {
    switch ($mode) {
        case 1:
            foreach ($xs as &$w) { $w++; }
            break;
        default:
            echo 'none';
    }
    unset($w);
    return $xs;
}

function leavesLoop(array $xs) {
    while (true) {
        if ($xs) {
            foreach ($xs as &$q) { $q++; }
        }
        break;
    }
    unset($q);

    do {
        switch (count($xs)) {
            case 0:
                foreach ($xs as &$r) { $r--; }
                break;
        }
        break;
    } while (true);
    return $xs;
}

function byRefParameter(array $xs, &$item) {
    foreach ($xs as $item) {}
    unset($item);
    $item = 'local only';
}

function sharedBindings(array $xs, array &$out) {
    global $shared;
    static $kept;
    foreach ($xs as $shared) {}
    unset($shared);
    foreach ($xs as $kept) {}
    unset($kept);

    $slot = &$out[0];
    foreach ($xs as $slot) {}
    unset($slot);

    $alias = &$source;
    foreach ($xs as $source) {}
    unset($source);

    $pair = [&$cell];
    foreach ($xs as $cell) {}
    unset($cell);

    $log = function () use (&$line) { echo $line; };
    foreach ($xs as $line) {}
    unset($line);
}

function laterBindingInLoop(array $groups) {
    foreach ($groups as $group) {
        foreach ($group as $n) {}
        unset($n);
        foreach ($group as &$n) { $n++; }
    }
    return $groups;
}

$collect = function (array $xs) use (&$acc) {
    foreach ($xs as $acc) {}
    unset($acc);
};
