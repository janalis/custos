<?php
/** @custos-protected */ function openGate() {}
function nestedPeerCheck($expected) {
    if ($_SERVER['REMOTE_ADDR'] === '192.0.2.10') {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function nestedBuiltinOnPeer($expected, array $proxies) {
    if (in_array($_SERVER['REMOTE_ADDR'], $proxies, true)) {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function nestedMethodCheck($expected, $request) {
    if ($request->isFromTrustedProxy()) {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function elseArm($expected) {
    if (!Proxies::knows($_SERVER['REMOTE_ADDR'])) {
        http_response_code(403);
    } else {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function elseifCondition($expected, $maintenance) {
    if ($maintenance) {
        http_response_code(503);
    } elseif (knownProxy()) {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
function guardClause($expected) {
    if (!knownProxy($_SERVER['REMOTE_ADDR'])) { return; }
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function guardClauseThrows($expected) {
    if ($_SERVER['REMOTE_ADDR'] !== '192.0.2.10') { throw new RuntimeException('direct access'); }
    $attempt = 1;
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function guardInElse($expected) {
    if (knownProxy()) { $via = 'proxy'; } else { return; }
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function guardInElseif($expected, $cli) {
    if ($cli) { $via = 'cli'; } elseif (!knownProxy()) { return; }
    if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
}
function guardOfOuterList($expected, $enabled) {
    if (!knownProxy()) { return; }
    if ($enabled) {
        if ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) { openGate(); }
    }
}
