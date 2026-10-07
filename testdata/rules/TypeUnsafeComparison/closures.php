<?php
function sameHandler(\Closure $onDone, \Closure $handler, callable $cb, $other) {
    $a = $onDone == $handler;
    $b = $handler != $onDone;
    $c = <error descr="\Closure has no __toString(), so it cannot be compared to a string.">$onDone == 'done'</error>;
    $d = <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$cb == $other</weak_warning>;
    return [$a, $b, $c, $d];
}
