<?php
define('ROWS', count($GLOBALS['rows']));
class Box {
    public function __construct(public array $all = []) {}
    public function take($v) {}
    public static function keep($v) {}
}
function shapes(array $a, $o, $cls, $m, $stop) {
    for ($i = 0; $i < count($a); $i++, $stop++) { echo $a[$i]; }      // E1: two steps
    for ($i = 0; ; $i++) { echo $a[$i]; }                             // E3: no condition
    for ($i = 0; $i < count($a); $j++) { echo $a[$i]; }               // E1: other variable stepped
    for ($i = 0; check($i); $i++) { echo $a[$i]; }                     // E3: not binary
    for ($stop += 1, $o->p = 0, $i = 0; $i < count($a); $i++) {
        echo $a[$i];
    }
}
function chained(array $a) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0, $n = $m = count($a); $i < $n; $i++) {
        echo $a[$i];
    }
}
function defaulted(array $a, $n = 3) {
    for ($i = 0; $i < $n; $i++) { echo $a[$i]; }                       // E5: parameter default
}
function constLimit($n = ROWS) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $GLOBALS['rows'][$i]; }
}
function chainedStatement(array $a) {
    $m = $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $a[$i]; }
}
function unsetAfter($n = ROWS) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) { echo $GLOBALS['rows'][$i]; }
    unset($n);
}
function conditionAssigned(array $a) {
    if ($n = count($a)) {
        <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) {
            echo $a[$i];
        }
    }
}
function contexts(array $a, Box $box, $o, $m, $cls) {
    $x = [];
    $n = count($a);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) {
        $a[$i]->run();
        echo $x[$a[$i]];
        echo $a[$i][0][1];
        $y = $a[$i][0][1];
        if ($a[$i]) {} elseif ($a[$i]) {}
        switch ($a[$i]) { case $a[$i]: echo "same"; break; }
        for ($j = $a[$i]; $j < 2; $j++) {}
        for (; $a[$i];) { break; }
        while ($a[$i]) {}
        do {} while ($a[$i]);
        foreach ($a[$i] as $v) {}
        strlen($a[$i]);
        unknownFn($a[$i]);
        strlen(...$a[$i]);
        $o->$m($a[$i]);
        $o->take($a[$i]);
        $box->take($a[$i]);
        $box->missing($a[$i]);
        $cls::keep($a[$i]);
        Box::keep($a[$i]);
        Box::missing($a[$i]);
        new Box($a[$i]);
        if ($o) {
            return $a[$i];
        }
        bump($a[$i]);
    }
}
function bump(&$v) {}
<warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < ROWS; $i++) {
    echo $GLOBALS['rows'][$i];
}
