<?php
function ensureDirs($cache, $logs, $cfg) {
    if (!is_dir($logs) && <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($cache) && !is_dir(...)'.">!mkdir($cache)</error> && !is_dir($logs)) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($cache, 0700) || is_dir(...)'.">mkdir($cache, 0700)</error> || is_dir(dirname($cache))) {}

    if (!is_dir($cache) && !mkdir($cache) && !IS_DIR( $cache )) {}
    if (!mkdir($d = $cfg->path()) && !is_dir($d)) {}
    if (!mkdir($cfg->path()) && !is_dir($cfg->path())) {}
}
