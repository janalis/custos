<?php
declare(ticks=1);
declare(ticks=1, strict_types=1);

namespace App;

use function call_user_func_array;

function measure(array $args) {
    return strlen(...$args);
}
