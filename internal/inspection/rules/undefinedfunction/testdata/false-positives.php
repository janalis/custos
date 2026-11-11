<?php
namespace App;

// Functions of PHP and of extensions the stubs describe, at any version.
apcu_fetch('key');
each([]);
mysql_connect();
\Swoole\Coroutine\run(function () {});

// A namespace an extension uses may hold functions the stubs lack.
\MongoDB\BSON\fromJSON('{}');
// A leading word no builtin function uses: an extension the stubs lack.
newext_open('handle');

// Guarded by name anywhere in the file.
if (!function_exists('App\optionalHelper')) {
    return;
}
optionalHelper();
if (is_callable('legacyRender')) {
    \legacyRender();
}

// Inside a condition about existence, extensions or versions.
if (extension_loaded('acme')) {
    acmeStart();
} elseif (PHP_VERSION_ID >= 90000) {
    futureFunction();
}
$value = defined('ACME') ? acmeValue() : null;
version_compare(PHP_VERSION, '9.0', '>=') && futureOther();
class_exists(...);

// Declared conditionally, here or in another file.
if (!function_exists('App\polyfilled')) {
    function polyfilled() { helperOfPolyfill(); }
}
polyfilled();

// Guarded by an early exit, or inside a switch or match arm.
function nativeStart()
{
    if (!extension_loaded('acme')) {
        return;
    }
    \Acme\Native\start();
}
function futureStart()
{
    if (PHP_VERSION_ID < 90000) {
        return;
    }
    \Php9\start();
}
function laterStart()
{
    if (version_compare(PHP_VERSION, '9.0', '<') || phpversion() < 9) {
        return;
    }
    \Php9\later();
}
function pdoStart()
{
    if (phpversion('pdo_acme') === false) {
        return;
    }
    pdo_acme_open();
}
switch (true) {
    case defined('ACME'):
        \Acme\Switched\start();
        break;
}
$matched = match (true) {
    extension_loaded('acme') => \Acme\Matched\start(),
    default => null,
};

// Dynamic calls and names in strings.
$fn = 'undefined_by_string';
$fn();
call_user_func('undefined_callback');
