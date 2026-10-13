<?php
function cut($s, $i, $o, $lines) {
    // Reported, but the operands cannot be evaluated once instead of twice.
    $a = substr($s, $i++, <warning descr="The length 'strlen($s) - $i++' is unnecessary; remove it.">strlen($s) - $i++</warning>);
    $b = substr(array_shift($lines), 0, <warning descr="Pass '-1' as the length instead.">strlen(array_shift($lines)) - 1</warning>);
    $c = substr($o->next(), 2, <warning descr="The length 'strlen($o->next()) - 2' is unnecessary; remove it.">strlen($o->next()) - 2</warning>);
    // Pure builtins are repeatable.
    $d = substr(trim($s), 0, <warning descr="Pass '-2' as the length instead.">strlen(trim($s)) - 2</warning>);
    $e = substr($s, count($lines), <warning descr="The length 'strlen($s) - count($lines)' is unnecessary; remove it.">strlen($s) - count($lines)</warning>);
    return [$a, $b, $c, $d, $e];
}
