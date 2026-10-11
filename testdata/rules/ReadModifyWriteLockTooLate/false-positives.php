<?php
function lockFileSpansCounter(string $path): void {
    $lock = fopen($path . '.lock', 'c');
    if ($lock === false || !flock($lock, LOCK_EX)) { return; }
    try {
        $n = (int) file_get_contents($path);
        file_put_contents($path, $n + 1, LOCK_EX);
    } finally {
        flock($lock, LOCK_UN);
        fclose($lock);
    }
}
function failedLockLeaves(string $path, $lock): void {
    if (!flock($lock, LOCK_EX | LOCK_NB)) { fclose($lock); return; }
    file_put_contents($path, file_get_contents($path) . "\n", LOCK_EX);
    flock($lock, LOCK_UN);
}
function unknownOperation(string $path, $lock, int $operation): void {
    flock($lock, $operation);
    $n = (int) file_get_contents($path);
    file_put_contents($path, $n + 1, LOCK_EX);
}
function delegatedLock(string $path, $mutex): void {
    $mutex->lock();
    $n = (int) file_get_contents($path);
    file_put_contents($path, $n + 1, LOCK_EX);
    $mutex->unlock();
}
function separateDestination(string $source, string $target): void {
    $n = (int) file_get_contents($source);
    file_put_contents($target, $n + 1, LOCK_EX);
}
