<?php
// E6: a randomly named directory must not exist yet.
function fresh($root, $flag) {
    $dir = sys_get_temp_dir() . '/gpg' . bin2hex(random_bytes(8));
    if (!mkdir($dir, 0700)) {
        throw new RuntimeException('taken');
    }
    if (!mkdir($root . '/' . uniqid('', true))) {}

    $mixed = $flag ? $root : $root . mt_rand();
    if (!mkdir($mixed) && !is_dir($mixed)) {}
    if (!mkdir($param) && !is_dir($param)) {}
}
