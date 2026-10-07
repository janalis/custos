<?php
function cacheDir($base) {
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($base) && !is_dir(...)'.">!MKDIR($base)</error>) {
        throw new RuntimeException('no cache');
    }
    if (!\MkDir($base) && !Is_Dir($base)) {}
}
