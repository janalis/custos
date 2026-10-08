<?php
function migrate(string $path): bool
{
    $lock = $path . '.lock';
    if (!@mkdir($lock, 0770, true)) {
        return true; // another process holds the lock
    }
    run_migrations();
    rmdir($lock);
    return true;
}

function migrateOther(string $path, string $other): bool
{
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($path, 0770, true) && !is_dir(...)'.">!@mkdir($path, 0770, true)</error>) {
        return false;
    }
    rmdir($other);
    return true;
}

$lockDir = sys_get_temp_dir() . '/job.lock';
if (!mkdir($lockDir)) {
    exit(0);
}
run_job();
rmdir($lockDir);
function scratch(string $tag)
{
    $dir = sys_get_temp_dir() . '/update-' . $tag;
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
    download($tag, $dir);
    rmdir($dir);
}
if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($cacheDir) && !is_dir(...)'.">!mkdir($cacheDir)</error>) {
    exit(2);
}
