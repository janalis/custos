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
    foreach ($a as $iValue) {
        echo $iValue;
    }
}
function chained(array $a) {
    foreach ($a as $iValue) {
        echo $iValue;
    }
}
function defaulted(array $a, $n = 3) {
    for ($i = 0; $i < $n; $i++) { echo $a[$i]; }                       // E5: parameter default
}
function constLimit($n = ROWS) {
    foreach ($GLOBALS['rows'] as $iValue) { echo $iValue; }
}
function chainedStatement(array $a) {
    $m = $n = count($a);
    foreach ($a as $iValue) { echo $iValue; }
}
function unsetAfter($n = ROWS) {
    foreach ($GLOBALS['rows'] as $iValue) { echo $iValue; }
    unset($n);
}
function conditionAssigned(array $a) {
    if ($n = count($a)) {
        foreach ($a as $iValue) {
            echo $iValue;
        }
    }
}
function contexts(array $a, Box $box, $o, $m, $cls) {
    $x = [];
    foreach ($a as $i => $iValue) {
        $iValue->run();
        echo $x[$iValue];
        echo $iValue[0][1];
        $y = $a[$i][0][1];
        if ($iValue) {} elseif ($iValue) {}
        switch ($iValue) { case $iValue: echo "same"; break; }
        for ($j = $iValue; $j < 2; $j++) {}
        for (; $iValue;) { break; }
        while ($iValue) {}
        do {} while ($iValue);
        foreach ($iValue as $v) {}
        bump($a[$i]);
        strlen($iValue);
        unknownFn($a[$i]);
        strlen(...$a[$i]);
        $o->$m($a[$i]);
        $o->take($a[$i]);
        $box->take($iValue);
        $box->missing($a[$i]);
        $cls::keep($a[$i]);
        Box::keep($iValue);
        Box::missing($a[$i]);
        new Box($a[$i]);
        return $a[$i];
    }
}
function bump(&$v) {}
foreach ($GLOBALS['rows'] as $iValue) {
    echo $iValue;
}
