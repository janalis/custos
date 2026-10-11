<?php
declare(strict_types=0);

namespace App;

function measure(array $args) {
    return strlen(...$args);
}
