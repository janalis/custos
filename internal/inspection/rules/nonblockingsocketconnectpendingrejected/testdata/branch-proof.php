<?php
function conditionalRejection($s) {
socket_set_nonblock($s);
if (!($result = socket_connect($s, "localhost", 80))) { if ($result === false) { throw new RuntimeException(); } }
}
function successfulValue($s) {
socket_set_nonblock($s);
if (!!socket_connect($s, "localhost", 80)) { return false; }
}
function rejectEitherPath($s, $flag) {
socket_set_nonblock($s);
if (!<warning descr="Handle a pending nonblocking connection.">socket_connect($s, "localhost", 80)</warning>) { if ($flag) { return false; } else { throw new RuntimeException(); } }
}
function helperHandlesFailure($s) {
    socket_set_nonblock($s);
    if (!socket_connect($s, "localhost", 80)) { handlePending($s); return false; }
}
