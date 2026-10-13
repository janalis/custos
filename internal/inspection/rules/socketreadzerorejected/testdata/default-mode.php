<?php
function readChunk($socket) {
    if (!($chunk = <warning descr="Compare socket read failure strictly with false.">socket_read($socket, 4096)</warning>)) { return false; }
    return $chunk;
}
function readNamed($socket) {
    if (<warning descr="Compare socket read failure strictly with false.">socket_read(length: 4096, socket: $socket)</warning> == false) { throw new RuntimeException('closed'); }
}
function readNamedMode($socket) {
    if (!<warning descr="Compare socket read failure strictly with false.">socket_read($socket, 4096, mode: PHP_BINARY_READ)</warning>) { throw new RuntimeException('closed'); }
}
