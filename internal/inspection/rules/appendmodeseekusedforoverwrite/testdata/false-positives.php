<?php
function replaceLocked($path, $data)
{
    $fp = fopen($path, 'a+');
    if ($fp === false) {
        return;
    }
    flock($fp, LOCK_EX);
    ftruncate($fp, 0);
    if (fseek($fp, 0) !== 0) {
        return;
    }
    fwrite($fp, $data);
    fclose($fp);
}

function maybeReset($path, $data, $reset, $size)
{
    $fp = fopen($path, 'a+');
    if ($fp === false) {
        return;
    }
    if ($reset) {
        ftruncate($fp, $size);
    }
    if (!rewind($fp)) {
        return;
    }
    fputs($fp, $data);
}
