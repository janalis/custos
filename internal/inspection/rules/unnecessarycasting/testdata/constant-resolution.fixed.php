<?php
namespace EngineConstants;

use const EngineConstants\TEXT as PHP_VERSION;
use const PHP_INT_MAX as LIMIT;

const PHP_INT_MAX = 'local limit';
const PHP_EOL = 42;
const TEXT = 'local version';

function readConstants() {
    return [
        PHP_INT_MAX,
        (int) PHP_INT_MAX,
        PHP_EOL,
        (string) PHP_EOL,
        PHP_VERSION,
        LIMIT,
        \PHP_INT_MAX,
        \PHP_EOL,
        (int) php_int_max,
        (string) namespace\PHP_VERSION,
    ];
}
