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
    if (!@mkdir($path, 0770, true) && !is_dir($path)) {
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
    if (!mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
    download($tag, $dir);
    rmdir($dir);
}
if (!mkdir($cacheDir) && !is_dir($cacheDir)) {
    exit(2);
}
