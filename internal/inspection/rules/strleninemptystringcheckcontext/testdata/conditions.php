<?php
function conditions(string $s, $x) {
    if ($x) {} elseif (<weak_warning descr="Compare with an empty string instead: '$s !== '''.">strlen($s)</weak_warning>) {}
    while (<weak_warning descr="Compare with an empty string instead: '$s !== '''.">strlen($s)</weak_warning>) { break; }
    do {} while (<weak_warning descr="Compare with an empty string instead: '$s !== '''.">strlen($s)</weak_warning>);
}
