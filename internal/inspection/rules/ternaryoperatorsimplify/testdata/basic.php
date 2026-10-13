<?php
function probe($n, $m, $a, $b, $obj) {
    $r = <weak_warning descr="Replace the ternary with '$n !== 7'.">$n !== 7 ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($n <= 7)'.">$n <= 7 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '(bool)($m % 2)'.">$m % 2 ? TRUE : FALSE</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($m | 4)'.">$m | 4 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '($a || $b)'.">$a || $b ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($a || $b)'.">$a || $b ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '(bool)($obj instanceof Countable)'.">$obj instanceof Countable ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$n == $m'.">($n == $m) ? (true) : (false)</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$n == $m'.">$n<>$m ? \false : \true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($a and $b)'.">($a and $b) ? false : true</weak_warning>;
    $r = (<weak_warning descr="Replace the ternary with '$a < $b'.">$a < $b ? true : false</weak_warning>);
    return $r;
}
