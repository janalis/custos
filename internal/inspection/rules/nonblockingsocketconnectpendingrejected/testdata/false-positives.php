<?php

function isPending($s)
{
    return in_array(socket_last_error($s), [SOCKET_EINPROGRESS, SOCKET_EALREADY], true);
}

// The helper receives the socket and may accept the pending state.
function delegatesToHelper($s, $host)
{
    socket_set_nonblock($s);
    if (!socket_connect($s, $host, 443)) {
        if (isPending($s)) {
            return true;
        }
        return false;
    }
    return true;
}

class Connector
{
    private static $shared;

    public function delegatesToMethod($s, $host)
    {
        socket_set_nonblock($s);
        if (!socket_connect($s, $host, 443)) {
            if ($this->pending($s)) {
                return true;
            }
            return false;
        }
        return true;
    }

    public function delegatesToStatic($s, $host)
    {
        socket_set_nonblock($s);
        if (!socket_connect($s, $host, 443)) {
            if (Errors::pending($s)) {
                return true;
            }
            return false;
        }
        return true;
    }

    // A static property is not a tracked socket.
    public function sharedSocket($host)
    {
        if (socket_set_nonblock(self::$shared)) {
            if (!socket_connect(self::$shared, $host, 443)) {
                report($host);
                return false;
            }
        }
        return true;
    }
}

function readsSocketOption($s, $host)
{
    socket_set_nonblock($s);
    if (!socket_connect($s, $host, 443)) {
        if (socket_get_option($s, SOL_SOCKET, SO_ERROR) !== 0) {
            return false;
        }
    }
    return true;
}

// The switch to nonblocking mode is not proven to have run.
function conditionalSwitch($s, $host, $async)
{
    if ($async) {
        socket_set_nonblock($s);
    }
    if (!socket_connect($s, $host, 443)) {
        return false;
    }
    return true;
}

function switchedBack($s, $host)
{
    socket_set_nonblock($s);
    socket_set_block($s);
    if (!socket_connect($s, $host, 443)) {
        return false;
    }
    return true;
}

function silencedSuccessBranch($s, $host)
{
    socket_set_nonblock($s);
    if (@socket_connect($s, $host, 443)) {
        return false;
    }
    return true;
}
