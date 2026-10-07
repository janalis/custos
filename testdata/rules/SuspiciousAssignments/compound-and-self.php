<?php
function still_reported($c, $d) {
    if ($c) { $acc .= 'x'; $acc = 'y'; }
    <error descr="$acc is overwritten right after the 'if'; an 'else' may be missing.">$acc = 'z'</error>;
    $same = 1;
    <error descr="$same is overwritten right after being assigned.">$same</error> = $d;
    return [$acc, $same];
}
