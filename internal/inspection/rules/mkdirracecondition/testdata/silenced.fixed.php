<?php
function prepare(string $cache, string $logs): void
{
    if (!@mkdir($cache, 0775, true) && !is_dir($cache)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $cache)); }
    if (!@mkdir($logs) && !is_dir($logs)) {
        echo 'failed';
    }
}
