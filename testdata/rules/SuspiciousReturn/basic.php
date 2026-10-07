<?php
function load($path) {
    try {
        if (!is_file($path)) {
            throw new InvalidArgumentException($path);
        }
        $data = file($path);
    } finally {
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return null;</error>
    }
}

function fetch($id) {
    try {
        foreach ([$id] as $k) { return $k; }
    } catch (Exception $e) {
    } finally {
        $cleanup = function () { return true; };
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return;</error>
    }
}

function quiet() {
    try {
        work();
    } catch (Exception $e) {
        return 1;
    } finally {
        return 0;
    }
}
