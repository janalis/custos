<?php
function cacheDir(string $dir): bool
{
    if (!is_dir($dir)) {
        @mkdir($dir, 0777, true);
    }
    if (!is_dir($dir)) @mkdir($dir);
    return true;
}

function otherGuard(string $dir, string $log): void
{
    if (!is_dir($log)) {
        if (!@mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
    }
}

function loud(string $dir): void
{
    if (!is_dir($dir)) {
        if (!mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); }
    }
}
