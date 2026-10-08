<?php
function thumbs(string $dir): bool
{
    if (!is_dir(dirname($dir)) || <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($dir) && !is_dir(...)'.">!@mkdir($dir)</error>) {
        return false;
    }
    if (is_writable($dir) && <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($dir) || is_dir(...)'.">mkdir($dir)</error>) {
        return true;
    }
    $ok = is_dir($dir) or <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($dir) || is_dir(...)'.">mkdir($dir)</error>;
    return $ok;
}
