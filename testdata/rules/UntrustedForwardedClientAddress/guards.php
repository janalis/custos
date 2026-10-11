<?php
/** @custos-protected */ function openGate() {}
function flagOnly($expected, $enabled) {
    if ($enabled) {
        if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) { openGate(); }
    }
}
function builtinGuard($expected) {
    if (!is_string($expected)) { return; }
    if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$expected == $_SERVER['HTTP_X_FORWARDED_FOR']</warning>) { openGate(); }
}
function checkFallsThrough($expected) {
    if (!knownProxy($_SERVER['REMOTE_ADDR'])) { error_log('unknown proxy'); }
    if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) { openGate(); }
}
function guardOutsideClosure($expected) {
    if (!knownProxy($_SERVER['REMOTE_ADDR'])) { return null; }
    return function () use ($expected) {
        if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $expected</warning>) { openGate(); }
    };
}
