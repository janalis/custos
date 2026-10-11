<?php
function varyFirst() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    header('Access-Control-Allow-Origin: ' . $origin);
    header('Cache-Control: public, max-age=60');
}
function varyFirstAmongOthers() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Accept-Encoding, Origin');
    header('HTTP/1.1 200 OK');
    header('X-Frame-Options: DENY');
    header('Access-Control-Allow-Origin: ' . $origin);
    header('Cache-Control: public, max-age=60');
}
function varyAppendedBefore() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    header('Vary: Accept', false);
    header('Access-Control-Allow-Origin: ' . $origin);
    header('Cache-Control: public, max-age=60');
}
function varyAppendedAfter() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    header('Access-Control-Allow-Origin: ' . $origin);
    header('Vary: Accept', false);
    header('Cache-Control: public, max-age=60');
}
function unknownHeaderBefore($line) {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header($line);
    header('Access-Control-Allow-Origin: ' . $origin);
    header('Cache-Control: public, max-age=60');
}
