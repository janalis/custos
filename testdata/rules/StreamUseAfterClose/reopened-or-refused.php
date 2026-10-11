<?php
// A closure holding the variable by reference can store a new stream in it.
function rotate($first, $second)
{
    $h = fopen($first, 'wb');
    if ($h === false) {
        return false;
    }
    $reopen = function () use (&$h, $second) {
        $h = fopen($second, 'wb');
    };
    fclose($h);
    $reopen();
    return fwrite($h, 'next');
}

// fclose returning false refused to close: the stream is still open there.
function finish($path)
{
    $h = fopen($path, 'wb');
    if ($h === false) {
        return;
    }
    if (!fclose($h)) {
        fwrite($h, 'retry');
    }
}

function finishLater($path)
{
    $h = fopen($path, 'wb');
    if ($h === false) {
        return;
    }
    $closed = fclose($h);
    if ($closed === false) {
        fflush($h);
    }
}

function finishOrLeave($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return null;
    }
    if (fclose($h)) {
        return null;
    }
    return fgets($h);
}

$log = fopen('app.log', 'ab');
if ($log !== false) {
    $closed = fclose($log);
    if (!$closed) {
        fwrite($log, 'still open');
    }
}
