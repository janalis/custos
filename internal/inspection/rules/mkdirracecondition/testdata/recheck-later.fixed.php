<?php
function ensureWritable(string $path): bool
{
    if (!is_dir($path)) {
        mkdir($path, 0777, true);
    }
    return is_dir($path) && is_writable($path);
}

function ensureCache(string $root): void
{
    $dir = $root . '/cache';
    mkdir($dir);
    if (!is_dir($dir)) {
        throw new RuntimeException('no cache');
    }
}

function otherDir(string $a, string $b): bool
{
    if (!is_dir($a)) {
        if (!mkdir($a) && !is_dir($a)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $a)); }
    }
    return is_dir($b);
}

function laterClosure(string $a): callable
{
    if (!mkdir($a) && !is_dir($a)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $a)); }
    return function () use ($a) { return 1; };
}

function resolvedLater(string $p): ?string
{
    @mkdir($p);
    clearstatcache(true);
    $real = realpath($p);
    return $real === false ? null : $real;
}

function existsLater(string $p): bool
{
    mkdir($p, 0700);
    return file_exists($p) && IS_WRITABLE($p);
}

function reassigned(string $base): void
{
    $dir = $base;
    if (!file_exists($dir)) {
        if (!mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
    }
    $dir .= '/attachments';
    if (!file_exists($dir)) {
        if (!mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
    }
}
