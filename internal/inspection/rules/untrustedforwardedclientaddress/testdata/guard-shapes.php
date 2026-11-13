<?php
/** @custos-protected */ function openGate() {}
function shortCircuitExit($expected) {
    knownProxy($_SERVER['REMOTE_ADDR']) or exit;
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function shortCircuitThrow($expected) {
    Proxies::knows($_SERVER['REMOTE_ADDR']) || throw new RuntimeException('direct access');
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function storedNegatedExit($expected) {
    $trusted = knownProxy($_SERVER['REMOTE_ADDR']);
    if (!$trusted) { return; }
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function storedElseArm($expected) {
    $trusted = knownProxy($_SERVER['REMOTE_ADDR']);
    if (!$trusted) {
        http_response_code(403);
    } else {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function storedFallsThrough($expected) {
    $trusted = knownProxy($_SERVER['REMOTE_ADDR']);
    if (!$trusted) { error_log('unknown proxy'); }
    if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) { openGate(); }
}
function shortCircuitLogs($expected) {
    knownProxy($_SERVER['REMOTE_ADDR']) or error_log('unknown proxy');
    if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) { openGate(); }
}
