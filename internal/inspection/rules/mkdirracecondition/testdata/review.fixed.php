<?php
function named($file, $mode, $o, $args) {
    if (!mkdir(permissions: $mode, directory: $file) && !is_dir($file)) {}
    if (!mkdir(permissions: $mode, directory: $file) && !is_dir($file)) {}
    mkdir(...$args);
    mkdir($file, ...$args);
    mkdir(permissions: $mode);
}

function comparisons($d, $e, $dry) {
    if (!mkdir($d) && !is_dir($d)) {}
    if (!mkdir($d) && !is_dir($d)) {}
    if (!mkdir($d) && !is_dir($d)) {}
    if (mkdir($d) || is_dir($d)) {}
    if (mkdir($d) || is_dir($d)) {}
    if (mkdir($d) xor $dry) {}
    if (mkdir($d) === null) {}
    if (mkdir($d) == $dry) {}
    if (!mkdir($d) & !mkdir($e)) {}
    if (mkdir($d) . 'x') {}
}

function strings($base, $seq, $name) {
    if (!mkdir("{$base}/$name") && !is_dir("{$base}/$name")) { throw new \RuntimeException(sprintf('Directory "%s" was not created', "{$base}/$name")); }
}
