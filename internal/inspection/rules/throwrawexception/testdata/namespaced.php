<?php
namespace Billing;

use Exception;
use Exception as Base;

function f($n) {
    if ($n) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Exception</weak_warning>('declined');
    }
    if ($n > 1) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Base</weak_warning>('aliased');
    }
    throw new Exception\Declined('card');
}
