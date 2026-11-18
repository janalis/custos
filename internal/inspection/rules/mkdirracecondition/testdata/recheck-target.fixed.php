<?php
function ensureDirs($cache, $logs, $cfg) {
    if (!is_dir($logs) && !mkdir($cache) && !is_dir($cache) && !is_dir($logs)) {}
    if (mkdir($cache, 0700) || is_dir(dirname($cache))) {}

    if (!is_dir($cache) && !mkdir($cache) && !IS_DIR( $cache )) {}
    if (!mkdir($d = $cfg->path()) && !is_dir($d)) {}
    if (!mkdir($cfg->path()) && !is_dir($cfg->path())) {}
}
