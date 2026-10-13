<?php
function closedBranch($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    if (fclose($h)) {
        <error descr="Use an open stream for this operation.">fread($h, 1)</error>;
    }
}

function closedAfterGuard($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    if (!fclose($h)) {
        return;
    }
    <error descr="Use an open stream for this operation.">rewind($h)</error>;
}

function storedSuccess($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    $closed = fclose($h);
    if ($closed) {
        <error descr="Use an open stream for this operation.">ftell($h)</error>;
    }
}

function overwrittenResult($path, $flag)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    $closed = fclose($h);
    $closed = $flag;
    if (!$closed) {
        <error descr="Use an open stream for this operation.">fgets($h)</error>;
    }
}

function capturedByValue($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    $peek = function () use ($h) {
        return $h;
    };
    fclose($h);
    $peek();
    <error descr="Use an open stream for this operation.">fseek($h, 0)</error>;
}

function otherVariableCaptured($path)
{
    $h = fopen($path, 'rb');
    if ($h === false) {
        return;
    }
    $count = 0;
    $bump = function () use (&$count) {
        $inner = function () use (&$h) {
        };
        $count++;
    };
    fclose($h);
    $bump();
    <error descr="Use an open stream for this operation.">fflush($h)</error>;
}

$out = fopen('out.txt', 'wb');
if ($out !== false) {
    fclose($out);
    <error descr="Use an open stream for this operation.">fwrite($out, 'late')</error>;
}
