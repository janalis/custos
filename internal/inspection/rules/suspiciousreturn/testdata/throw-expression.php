<?php
function d($x) {
    try {
        $y = $x ?? throw new LogicException();
    } finally {
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return $y ?? null;</error>
    }
}
