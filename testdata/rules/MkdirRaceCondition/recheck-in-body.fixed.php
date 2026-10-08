<?php
function ensure_dir(string $dir): void
{
    if (!@mkdir($dir, 0777, true)) {
        clearstatcache();
        if (is_dir($dir)) {
            return; // created by another process meanwhile
        }
        throw new RuntimeException('cannot create ' . $dir);
    }
}

function ensure_other(string $dir, string $log): void
{
    if (!mkdir($dir) && !is_dir($dir)) {
        if (is_dir($log)) {
            return;
        }
    }
}
