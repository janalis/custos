<?php
function sharedLockDoesNotSerialise(string $path, $lock): void {
    flock($lock, LOCK_SH);
    $n = (int) file_get_contents($path);
    <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
}
function releasedBeforeRead(string $path, $lock): void {
    flock($lock, LOCK_EX);
    flock($lock, LOCK_UN);
    $n = (int) file_get_contents($path);
    <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
}
function releasedBeforeWrite(string $path, $lock): void {
    flock($lock, LOCK_EX);
    $n = (int) file_get_contents($path);
    fclose($lock);
    <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
}
function lockedAfterRead(string $path, $lock): void {
    $n = (int) file_get_contents($path);
    flock($lock, LOCK_EX);
    <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
    flock($lock, LOCK_UN);
}
function releasedAroundWrite(string $path, $lock): void {
    flock($lock, LOCK_EX);
    $n = (int) file_get_contents($path);
    if (flock($lock, LOCK_UN)) {
        <warning descr="Hold the lock across the complete read-modify-write operation.">file_put_contents($path, $n + 1, LOCK_EX)</warning>;
    }
}
