<?php
function earlyExit(string $host, int $port)
{
    $s = fsockopen($host, $port, $errno, $errstr, 2);
    if ($s === false || !stream_set_timeout($s, 1)) {
        return null;
    }
    fwrite($s, "GET / HTTP/1.0\r\n\r\n");
    $line = fgets($s);
    return [$line, stream_get_meta_data($s)['timed_out']];
}

function earlyThrow(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2.5);
    if ($s === false || stream_set_timeout($s, 5) === false) {
        throw new RuntimeException($errstr);
    }
    return fgets($s);
}

function elseAfterFailure(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    if (!$s || !stream_set_timeout($s, 5)) {
        return null;
    } else {
        return fread($s, 10);
    }
}

function shortCircuit(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 2);
    return $s && stream_set_timeout($s, 5) ? fgets($s) : null;
}

function zeroFloat(string $host)
{
    $s = fsockopen($host, 80, $errno, $errstr, 0.0);
    return fgets($s);
}
