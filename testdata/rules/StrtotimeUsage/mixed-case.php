<?php
function stamps($spec) {
    $a = <warning descr="Call time() instead of parsing 'now'.">StrToTime('now')</warning>;
    $b = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">STRTOTIME($spec, Time())</warning>;
    return [$a, $b];
}
