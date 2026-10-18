<?php
function conditionalRejection($s) {

if (!($result = socket_read($s, 32))) { if ($result === false) { throw new RuntimeException(); } }
}
function successfulValue($s) {

if (!!socket_read($s, 32)) { return false; }
}
function rejectEitherPath($s, $flag) {

if (!<warning descr="Compare socket read failure strictly with false.">socket_read($s, 32)</warning>) { if ($flag) { return false; } else { throw new RuntimeException(); } }
}
