<?php
function nonAdjacentGuard(string $path, string $manifest): int
{
    $digest = hash_file('sha256', $path);
    if ($digest === false) {
        return -1;
    }
    $line = basename($path);
    return (int) file_put_contents($manifest, $digest);
}

function truthyGuard(string $path, string $manifest): int
{
    $digest = hash_file('sha256', $path);
    if (!$digest) {
        return -1;
    }
    return (int) file_put_contents($manifest, $digest);
}

function shortCircuit(string $path): bool
{
    $digest = hash_file('sha256', $path);
    return $digest !== false && $digest === 'abc';
}

function combinedGuard(string $path, string $manifest): int
{
    $digest = hash_file('sha256', $path);
    if ($digest === false || $manifest === '') {
        return -1;
    }
    return (int) file_put_contents($manifest, $digest);
}

function truthyBranch(string $path, string $manifest): void
{
    $digest = hash_file('sha256', $path);
    if ($digest) {
        file_put_contents($manifest, $digest);
    }
}

function ternaryGuard(string $path): bool
{
    $digest = hash_file('sha256', $path);
    return $digest ? $digest === 'abc' : false;
}
