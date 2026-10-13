<?php
declare(strict_types=1);

namespace App;

function width(string $s): int { return \strlen($s); }

function measure(array $args) {
    // Bound at run time: the callback still coerces its arguments, no fix.
    $a = call_user_func_array('App\width', $args);
    $b = call_user_func_array('strlen', $args);
    // Bound at compile time: already strict, fixed.
    $c = strlen(...$args);
}
