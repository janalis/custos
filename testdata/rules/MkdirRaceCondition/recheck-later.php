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
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($a) && !is_dir(...)) { ... }'.">mkdir($a);</error>
    }
    return is_dir($b);
}

function laterClosure(string $a): callable
{
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($a) && !is_dir(...)) { ... }'.">mkdir($a);</error>
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
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
    }
    $dir .= '/attachments';
    if (!file_exists($dir)) {
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
    }
}
