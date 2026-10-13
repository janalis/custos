<?php
function search(array $a, $x) {
    for ($i = 0; $i < count($a); $i++) {        // D8c: counter read after the loop
        if ($a[$i] === $x) {
            break;
        }
    }
    if ($i == count($a)) {
        return -1;
    }
    return $i;
}
function limitAfter(array $a) {
    for ($i = 0, $n = count($a); $i < $n; $i++) {   // D8c: header limit read after
        echo $a[$i];
    }
    echo $n;
}
function reassignedInBranch(array $a, $f) {
    for ($i = 0; $i < count($a); $i++) {        // a conditional re-assignment does not dominate
        echo $a[$i];
    }
    if ($f) {
        $i = 0;
    }
    echo $i;
}
function selfRead(array $a) {
    for ($i = 0; $i < count($a); $i++) {        // the re-assignment reads the old value
        echo $a[$i];
    }
    $i = $i + 1;
}
function nested(array $a) {
    for ($i = 0; $i < count($a); $i++) {        // re-assigned inside a call
        echo $a[$i];
    }
    reset($i = 0);
}
function compound(array $a) {
    for ($i = 0; $i < count($a); $i++) {
        echo $a[$i];
    }
    $i += 1;
}
function capture(array $a) {
    for ($i = 0; $i < count($a); $i++) {
        echo $a[$i];
    }
    return function () use ($i) { return $i; };
}
function notAfter(array $a) {
    $n = 3;
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($a); $i++) {
        echo $a[$i];
    }
    echo $n;                                    // a limit not assigned in the header
    function inner() { return $i; }             // another scope
}
function reassigned(array $a, array $b) {
    for ($i = 0; $i < count($a); $i++) {        // next mention is a conditional loop init: kept conservative
        echo $a[$i];
    }
    if ($b) {
        <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < count($b); $i++) {
            echo $b[$i];
        }
    }
    $i = 10;
    echo $i;
}
switch ($mode) {
    case 1:
        <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($k = 0; $k < count($GLOBALS['rows']); $k++) {
            echo $GLOBALS['rows'][$k];
        }
        $k = 0;
        break;
}
for ($j = 0; $j < count($GLOBALS['rows']); $j++) {      // top level: read after
    echo $GLOBALS['rows'][$j];
}
echo $j;
