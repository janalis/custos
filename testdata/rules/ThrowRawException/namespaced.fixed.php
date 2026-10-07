<?php
namespace Billing;

use Exception;
use Exception as Base;

function f($n) {
    if ($n) {
        throw new \RuntimeException('declined');
    }
    if ($n > 1) {
        throw new \RuntimeException('aliased');
    }
    throw new Exception\Declined('card');
}
