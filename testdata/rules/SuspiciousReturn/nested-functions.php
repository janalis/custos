<?php
function f($items) {
    try {
        $f = function () { return 1; };
        if (!$items) {
            return $f;
        }
    } finally {
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return null;</error>
    }
}
