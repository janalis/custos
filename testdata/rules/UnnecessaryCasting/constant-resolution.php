<?php
namespace EngineConstants;

use const EngineConstants\TEXT as PHP_VERSION;
use const PHP_INT_MAX as LIMIT;

const PHP_INT_MAX = 'local limit';
const PHP_EOL = 42;
const TEXT = 'local version';

function readConstants() {
    return [
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> PHP_INT_MAX,
        (int) PHP_INT_MAX,
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> PHP_EOL,
        (string) PHP_EOL,
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> PHP_VERSION,
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> LIMIT,
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> \PHP_INT_MAX,
        <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> \PHP_EOL,
        (int) php_int_max,
        (string) namespace\PHP_VERSION,
    ];
}
