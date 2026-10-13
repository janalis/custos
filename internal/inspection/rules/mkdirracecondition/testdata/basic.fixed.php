<?php
function prepare($root, $cfg) {

    if (!mkdir($root) && !is_dir($root)) {}
    if (mkdir("$root/log", 0700, true) || is_dir("$root/log")) {}
    if (mkdir($concurrentDirectory = dirname($root)) || is_dir($concurrentDirectory)) {}
    if (file_exists($root) || mkdir($root, permissions: 0700) || is_dir($root)) {}
    if ($cfg->ok && !mkdir($concurrentDirectory = strtolower($root)) && !is_dir($concurrentDirectory)) {}
    if (!mkdir($root) && !is_dir($root)) {}
}
function tempVar($cfg) {
    if (mkdir($concurrentDirectory = $cfg->dir(), 0700) || is_dir($concurrentDirectory)) { return 1; }
    return 0;
}
