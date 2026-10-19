<?php
function invertedBranch(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    if (!stream_set_timeout($s, 1)) {
        return <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>;
    }
    return null;
}

function elseBranch(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    if ($s && stream_set_timeout($s, 5)) {
        return null;
    } else {
        return <warning descr="Set a separate timeout for stream reads.">fread($s, 10)</warning>;
    }
}

function uncheckedStatement(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    stream_set_timeout($s, 5);
    return <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>;
}

function guardDoesNotExit(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    if (!stream_set_timeout($s, 5)) {
        error_log('no read timeout');
    }
    return <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>;
}

function floatTimeout(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2.5);
    return <warning descr="Set a separate timeout for stream reads.">fread($s, 10)</warning>;
}
function timeoutStoredThenReplaced($flag) {
    $s = fsockopen("example.test", 443, $errno, $errstr, 2);
    if ($s === false) { return null; }
    $ok = stream_set_timeout($s, 2);
    $ok = $flag;
    if (!$ok) { return null; }
    return <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>;
}
