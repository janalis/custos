<?php
function cacheDir($base) {
    if (!mkdir($base) && !is_dir($base)) {
        throw new RuntimeException('no cache');
    }
    if (!\MkDir($base) && !Is_Dir($base)) {}
}
