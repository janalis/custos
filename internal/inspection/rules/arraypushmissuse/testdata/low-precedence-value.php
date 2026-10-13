<?php
function flags(array $list, $x, $y) {
    <warning descr="Use '$list[] = ($x or $y)' instead; it avoids a function call.">array_push($list, $x or $y)</warning>;
    <warning descr="Use '$list[] = ($x and $y)' instead; it avoids a function call.">array_push($list, $x and $y)</warning>;
    <warning descr="Use '$list[] = ($x xor $y)' instead; it avoids a function call.">array_push($list, $x xor $y)</warning>;
    <warning descr="Use '$list[] = $x || $y' instead; it avoids a function call.">array_push($list, $x || $y)</warning>;
    <warning descr="Use '$list[] = $x ? 1 : 2' instead; it avoids a function call.">array_push($list, $x ? 1 : 2)</warning>;
    return $list;
}
