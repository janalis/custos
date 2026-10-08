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
    while (!is_dir($other) && <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($dir) && !is_dir(...)'.">!mkdir($dir)</error>) {
        usleep(1000);
    }
    $ok = !is_dir($dir) && <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($dir) && !is_dir(...)'.">!mkdir($dir)</error>;
    return $ok;
}
