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

// Dynamic calls and names in strings.
$fn = 'undefined_by_string';
$fn();
call_user_func('undefined_callback');
