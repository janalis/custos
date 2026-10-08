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
