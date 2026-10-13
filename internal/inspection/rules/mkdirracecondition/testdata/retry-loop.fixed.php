<?php
// A retry loop re-runs its is_dir() test on every iteration.
function ensure(string $dir, string $other) {
    $tries = 0;
    while (!is_dir($dir) && !@mkdir($dir, 0755, true)) {
        usleep(1000);
        if ($tries++ > 5) {
            throw new RuntimeException($dir);
        }
    }
    do {
        $tries--;
    } while (!is_dir($dir) && !mkdir($dir));
    while (!is_dir($other) && !mkdir($dir) && !is_dir($dir)) {
        usleep(1000);
    }
    $ok = !is_dir($dir) && !mkdir($dir) && !is_dir($dir);
    return $ok;
}
