<?php
function varyReplacedBefore() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    header('Vary: Accept');
    <warning descr="Include Vary: Origin for cacheable origin-dependent responses.">header('Access-Control-Allow-Origin: ' . $origin)</warning>;
    header('Cache-Control: public, max-age=60');
}
function varyReplacedAfter() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    <warning descr="Include Vary: Origin for cacheable origin-dependent responses.">header('Access-Control-Allow-Origin: ' . $origin)</warning>;
    header('Vary: Accept');
    header('Cache-Control: public, max-age=60');
}
function varyRemovedBefore() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('Vary: Origin');
    header_remove('Vary');
    <warning descr="Include Vary: Origin for cacheable origin-dependent responses.">header('Access-Control-Allow-Origin: ' . $origin)</warning>;
    header('Cache-Control: public, max-age=60');
}
function unrelatedHeadersBefore() {
    $origin = $_SERVER['HTTP_ORIGIN'];
    header('X-Frame-Options: DENY');
    <warning descr="Include Vary: Origin for cacheable origin-dependent responses.">header('Access-Control-Allow-Origin: ' . $origin)</warning>;
    header('Cache-Control: public, max-age=60');
}
