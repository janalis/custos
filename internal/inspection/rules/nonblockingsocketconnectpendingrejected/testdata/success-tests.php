<?php
// The nonblocking switch is proven by any test of its result guarding the connect.
function orExit($s, $host) {
    socket_set_nonblock($s) or exit(1);
    if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) { throw new RuntimeException(); }
}
function stored($s, $host) {
    $ok = socket_set_nonblock($s);
    if (!$ok) { return false; }
    if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, $host, 443)</warning>) { return false; }
    return true;
}
function storedThenReplaced($s, $host, $flag) {
    $ok = $flag && socket_set_nonblock($s);
    $ok = $flag;
    if ($ok) {
        if (!socket_connect($s, $host, 443)) { return false; }
    }
    return true;
}
