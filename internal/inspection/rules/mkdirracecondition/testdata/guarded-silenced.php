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
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">@mkdir($dir);</error>
    }
}

function loud(string $dir): void
{
    if (!is_dir($dir)) {
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
    }
}
