<?php
/** @custos-protected */ function openGate() {}
function plainEarlierArm($expected, $mode) {
    if ($mode === 'off') {
        return;
    } elseif (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) {
        openGate();
    }
}
function proxyFirstArm($expected) {
    if (!knownProxy($_SERVER['REMOTE_ADDR'])) {
        error_log('unknown proxy');
    } elseif ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) {
        openGate();
    }
}
function proxyEarlierElseif($expected, $mode) {
    if ($mode === 'off') {
        return;
    } elseif (!knownProxy($_SERVER['REMOTE_ADDR'])) {
        return;
    } elseif ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) {
        openGate();
    }
}
function proxyLaterArm($expected, $mode) {
    if ($mode === 'off') {
        return;
    } elseif (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) {
        openGate();
    } elseif (knownProxy($_SERVER['REMOTE_ADDR'])) {
        error_log('proxy');
    }
}
function enclosingGuard($expected, $mode) {
    if (!knownProxy($_SERVER['REMOTE_ADDR'])) { return; }
    if ($mode === 'off') {
        return;
    } elseif ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) {
        openGate();
    }
}
function unprotectedArm($expected, $mode) {
    if ($mode === 'off') {
        return;
    } elseif ($_SERVER['HTTP_X_FORWARDED_FOR'] === $expected) {
        error_log('match');
    }
}
