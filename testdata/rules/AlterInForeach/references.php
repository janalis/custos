<?php
function breaks(array $xs, $mode) {
    switch ($mode) {
        case 1:
            foreach ($xs as &<warning descr="Unset '$a' right after the loop: it is still a reference to the last element.">$a</warning>) { $a++; }
            break;
    }
    echo $mode;

    while ($mode) {
        while ($xs) {
            foreach ($xs as &<warning descr="Unset '$b' right after the loop: it is still a reference to the last element.">$b</warning>) { $b++; }
            break 2;
        }
    }
    unset($b);

    while ($mode) {
        foreach ($xs as &<warning descr="Unset '$c' right after the loop: it is still a reference to the last element.">$c</warning>) { $c++; }
        continue;
    }
    unset($c);

    foreach ($xs as &<warning descr="Unset '$d' right after the loop: it is still a reference to the last element.">$d</warning>) { $d++; };
    echo $mode;
}

function unrelatedBindings(array $xs, &$other, $plain) {
    global $shared;
    static $kept;
    $alias = &$xs[0];
    $inner = function (&$item) use (&$alias, $plain) { static $item2; return $item; };
    $arrow = fn (&$item) => $item;
    foreach ($xs as &$w) { $w++; }
    unset($w);
    foreach ($xs as $item) {}
    unset(<weak_warning descr="'$item' is not a reference here; unsetting it is unnecessary.">$item</weak_warning>);
    foreach ($xs as $plain) {}
    unset(<weak_warning descr="'$plain' is not a reference here; unsetting it is unnecessary.">$plain</weak_warning>);
}

function laterBinding(array $xs) {
    foreach ($xs as $n) {}
    unset(<weak_warning descr="'$n' is not a reference here; unsetting it is unnecessary.">$n</weak_warning>);
    foreach ($xs as &$n) { $n++; }
    unset($n);

    switch (count($xs)) {
        default:
            foreach ($xs as $m) {}
            unset(<weak_warning descr="'$m' is not a reference here; unsetting it is unnecessary.">$m</weak_warning>);
    }
    $m = &$xs[0];
    return $xs;
}

function earlierLoop(array $a, array $b) {
    foreach ($a as &<warning descr="Unset '$v' right after the loop: it is still a reference to the last element.">$v</warning>) { $v++; }
    foreach ($b as $v) {}
    unset($v);
    $v = 'tmp';
    return $a;
}

class Holder {
    public function run(array $xs) {
        $probe = new class {
            public function touch(&$entry) {}
        };
        function helper(&$entry) {}
        foreach ($xs as $entry) {}
        unset(<weak_warning descr="'$entry' is not a reference here; unsetting it is unnecessary.">$entry</weak_warning>);
    }
}
