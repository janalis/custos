<?php
function prepare(string $cache, string $logs): void
{
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($cache, 0775, true) && !is_dir(...)) { ... }'.">@mkdir($cache, 0775, true);</error>
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($logs) && !is_dir(...)'.">!@mkdir($logs)</error>) {
        echo 'failed';
    }
}
