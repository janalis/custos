<?php
namespace Billing;

function due($spec) {
    $now   = <warning descr="Call time() instead of parsing 'now'.">strtotime("Now")</warning>;
    $other = <warning descr="Call time() instead of parsing 'now'.">\strtotime('NOW')</warning>;
    $next  = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">strtotime($spec, \time())</warning>;
    $last  = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">strtotime('last monday', time())</warning>;
    $d = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">strtotime('now', time())</warning>;
    return [$now, $other, $next, $last, $d];
}
