<?php
declare(strict_types=1);

use Random\Engine\Mt19937;

$engines = [
    Mt19937::class,
    // the first import went before the code: a second one waits for the next pass
    'Random\Engine\Secure',
];
