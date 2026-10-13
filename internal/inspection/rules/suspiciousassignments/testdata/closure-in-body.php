<?php
function local_with_closure()
{
    $cb = function () { global $x; return $x; }; // another scope
    $x = 1;
    <error descr="$x is overwritten right after being assigned.">$x</error> = compute();
    return [$x, $cb];
}
