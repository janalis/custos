<?php
function lineMode($socket) {
    if (!socket_read($socket, 4096, PHP_NORMAL_READ)) { return false; }
    if (!socket_read($socket, 4096, mode: PHP_NORMAL_READ)) { return false; }
}
function unknownMode($socket, $mode, array $args) {
    if (!socket_read($socket, 4096, $mode)) { return false; }
    if (!socket_read(...$args)) { return false; }
    if (!socket_read($socket, ...$args)) { return false; }
}
function strictDefault($socket) {
    $chunk = socket_read($socket, 4096);
    if ($chunk === false) { throw new RuntimeException('failed'); }
    if (!socket_read($socket, 4096)) { echo 'nothing read'; }
    return $chunk;
}
function callableOnly() {
    if (!socket_read(...)) { return false; }
}
