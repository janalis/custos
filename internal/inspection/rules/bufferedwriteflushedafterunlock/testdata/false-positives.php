<?php

// A closure that captured the handle flushes it while the lock is held.
function closure_flush($h, string $data): void {
    $flush = function () use ($h) { fflush($h); };
    flock($h, LOCK_EX);
    fwrite($h, $data);
    $flush();
    flock($h, LOCK_UN);
    fflush($h);
}

// An arrow function shares the handle implicitly.
function arrow_flush($h, string $data): void {
    $flush = fn () => fflush($h);
    flock($h, LOCK_EX);
    fwrite($h, $data);
    $flush();
    flock($h, LOCK_UN);
    fflush($h);
}

// The handle is handed to other code in a non-stream slot.
function handed_over($h, string $data): void {
    flock($h, LOCK_EX);
    fwrite($h, $data);
    finish('log', ($h));
    flock($h, LOCK_UN);
    fflush($h);
}

function handed_to_constructor($h, string $data): void {
    flock($h, LOCK_EX);
    fwrite($h, $data);
    $writer = new Writer($h);
    flock($h, LOCK_UN);
    fflush($h);
}

// A variable variable names no single stream.
function variable_variable($a, $b, string $data): void {
    flock($$a, LOCK_EX);
    fwrite($$a, $data);
    flock($$a, LOCK_UN);
    fflush($$b);
}
