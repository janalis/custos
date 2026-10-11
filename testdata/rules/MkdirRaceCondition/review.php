<?php
function named($file, $mode, $o, $args) {
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir(permissions: $mode, directory: $file) && !is_dir(...)'.">!mkdir(permissions: $mode, directory: $file)</error>) {}
    if (!mkdir(permissions: $mode, directory: $file) && !is_dir($file)) {}
    mkdir(...$args);
    mkdir($file, ...$args);
    mkdir(permissions: $mode);
}

function comparisons($d, $e, $dry) {
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($d) && !is_dir(...)'.">mkdir($d) == false</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($d) && !is_dir(...)'.">false == mkdir($d)</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($d) && !is_dir(...)'.">mkdir($d) != true</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d) != false</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($d) || is_dir(...)'.">mkdir($d) == true</error>) {}
    if (mkdir($d) xor $dry) {}
    if (mkdir($d) === null) {}
    if (mkdir($d) == $dry) {}
    if (!mkdir($d) & !mkdir($e)) {}
    if (mkdir($d) . 'x') {}
}

function strings($base, $seq, $name) {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(&quot;{$base}/$name&quot;) && !is_dir(...)) { ... }'.">mkdir("{$base}/$name");</error>
}
