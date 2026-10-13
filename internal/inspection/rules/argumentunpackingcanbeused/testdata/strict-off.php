<?php
declare(strict_types=0);

namespace App;

function measure(array $args) {
    return <warning descr="Call 'strlen(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('strlen', $args)</warning>;
}
