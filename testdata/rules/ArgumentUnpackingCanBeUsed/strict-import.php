<?php
declare(ticks=1);
declare(ticks=1, strict_types=1);

namespace App;

use function call_user_func_array;

function measure(array $args) {
    return <warning descr="Call 'strlen(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('strlen', $args)</warning>;
}
