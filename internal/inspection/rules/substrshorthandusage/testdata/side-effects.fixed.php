<?php
function cut($s, $i, $o, $lines) {
    // Reported, but the operands cannot be evaluated once instead of twice.
    $a = substr($s, $i++, strlen($s) - $i++);
    $b = substr(array_shift($lines), 0, strlen(array_shift($lines)) - 1);
    $c = substr($o->next(), 2, strlen($o->next()) - 2);
    // Pure builtins are repeatable.
    $d = substr(trim($s), 0, -2);
    $e = substr($s, count($lines));
    return [$a, $b, $c, $d, $e];
}
