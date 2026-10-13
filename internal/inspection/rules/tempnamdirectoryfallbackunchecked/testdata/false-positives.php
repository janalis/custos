<?php

function storeSwapped(string $privateDir, string $secret): bool
{
    $path = tempnam($privateDir, 'job');
    if ($path === false) {
        return false;
    }
    if (realpath($privateDir) !== realpath(dirname($path))) {
        unlink($path);
        return false;
    }
    file_put_contents($path, $secret);
    return true;
}

function storeLoose(string $privateDir, string $secret): void
{
    $path = tempnam($privateDir, 'job');
    if ($path === false || realpath(dirname($path)) != realpath($privateDir)) {
        throw new RuntimeException('private directory unavailable');
    }
    file_put_contents($path, $secret);
}

function storeWhenMatching(string $privateDir, string $secret): bool
{
    $path = tempnam($privateDir, 'job');
    if ($path !== false && realpath(dirname($path)) === realpath($privateDir)) {
        file_put_contents($path, $secret);
        return true;
    }
    return false;
}

function scratch(string $data): string
{
    $path = tempnam(sys_get_temp_dir(), 'scratch');
    if ($path === false) {
        return '';
    }
    file_put_contents($path, $data);
    return $path;
}

function scratchHandle(string $data)
{
    $tmp = sys_get_temp_dir();
    $path = tempnam($tmp, 'scratch');
    return fopen($path, 'w');
}
