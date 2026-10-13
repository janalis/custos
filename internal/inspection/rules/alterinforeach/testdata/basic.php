<?php
function climb(array $list, $flag) {
    foreach ($list as &<warning descr="Unset '$a' right after the loop: it is still a reference to the last element.">$a</warning>) {}
    $a = 1;

    if ($flag) {
        foreach ($list as &$b) {}
    }
    unset($b);

    if ($flag) {
        foreach ($list as &<warning descr="Unset '$c' right after the loop: it is still a reference to the last element.">$c</warning>) {}
    } elseif ($list) {
        echo 1;
    }

    while ($flag) {
        foreach ($list as &<warning descr="Unset '$d' right after the loop: it is still a reference to the last element.">$d</warning>) {}
    }
    echo 2;

    if ($flag) foreach ($list as &<warning descr="Unset '$e' right after the loop: it is still a reference to the last element.">$e</warning>) {} else echo 3;

    foreach ($list as &$g) {}
    // plain comment
    unset($x, $g);

    foreach ($list as &$h) {}
    throw new Exception();
}

function ret(array $list) {
    foreach ($list as &$v) {}
    return $list;
}

function unsets(array $list) {
    foreach ($list as $k => $item) {}
    unset(<weak_warning descr="'$item' is not a reference here; unsetting it is unnecessary.">$item</weak_warning>, $k);

    foreach ($list as $one) {}
    /** doc */
    unset($one);

    foreach ($list as $three) {}
    unset($other);

    foreach ($list as [$x, $y]) {}
    unset($x);

    foreach ($list as $outer) {
        foreach ($list as $inner) {}
        echo 1;
    }
    unset(<weak_warning descr="'$inner' is not a reference here; unsetting it is unnecessary.">$inner</weak_warning>, <weak_warning descr="'$inner' is not a reference here; unsetting it is unnecessary.">$inner</weak_warning>);

    foreach ($list as $k => $q) {
        $list[$k] = $q;
    }
}

foreach ($list as &$top) {}
