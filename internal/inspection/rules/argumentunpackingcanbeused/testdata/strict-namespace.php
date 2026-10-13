<?php
declare(strict_types=1);

namespace App;

function width(string $s): int { return \strlen($s); }

function measure(array $args) {
    // Bound at run time: the callback still coerces its arguments, no fix.
    $a = <warning descr="Call '\App\width(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('App\width', $args)</warning>;
    $b = <warning descr="Call 'strlen(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('strlen', $args)</warning>;
    // Bound at compile time: already strict, fixed.
    $c = <warning descr="Call 'strlen(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">\call_user_func_array('strlen', $args)</warning>;
}
