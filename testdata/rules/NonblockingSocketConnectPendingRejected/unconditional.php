<?php

function connectAfterSwitch($s, $host)
{
    socket_set_nonblock($s);
    if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) {
        return false;
    }
    return true;
}

function connectSilenced($s, $host)
{
    socket_set_nonblock($s);
    if (!@<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) {
        throw new RuntimeException('connect failed');
    }
}

function connectLogsOnly($s, $host)
{
    socket_set_nonblock($s);
    if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) {
        error_log('connect to ' . $host . ' failed');
        socket_close($s);
        return false;
    }
    return true;
}
