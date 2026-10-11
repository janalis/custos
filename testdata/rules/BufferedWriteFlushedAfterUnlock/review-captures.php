<?php

// Nested scopes that do not share the handle leave the proof intact.
function unrelated_scopes($h, $other, string $data): void {
    $log = function () use ($other) { fflush($other); };
    $size = fn ($h) => strlen($data);
    $local = function ($h) { fflush($h); };
    flock($h, LOCK_EX);
    fwrite(($h), $data);
    $log();
    flock($h, LOCK_UN);
    <warning descr="Flush buffered writes before releasing the lock.">fflush($h)</warning>;
    report('done', $h);
}
