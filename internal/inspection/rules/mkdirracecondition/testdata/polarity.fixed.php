<?php
function thumbs(string $dir): bool
{
    if (!is_dir(dirname($dir)) || (!@mkdir($dir) && !is_dir($dir))) {
        return false;
    }
    if (is_writable($dir) && (mkdir($dir) || is_dir($dir))) {
        return true;
    }
    $ok = is_dir($dir) or mkdir($dir) || is_dir($dir);
    return $ok;
}
