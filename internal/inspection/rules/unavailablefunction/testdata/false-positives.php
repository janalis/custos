<?php
namespace App;

// Available at the target, in every version or since an older one.
strlen('x');
array_column([], 'id');
password_hash('secret', PASSWORD_DEFAULT);

// Removed before the target, not introduced after it.
mysql_connect();

// The New Relic agent versions its functions on its own.
newrelic_create_distributed_trace_payload();

// A namespaced function of the same name is the one called.
function str_starts_with($haystack, $needle) {}
str_starts_with('a', 'b');

// Declared by the project, guarded for PHP versions that have it.
str_ends_with('a', 'b');

// Guarded by name or by version.
if (function_exists('array_key_first')) {
    $first = array_key_first($items);
}
if (PHP_VERSION_ID >= 80000) {
    $found = str_contains($haystack, 'needle');
}
if (!function_exists('get_debug_type')) {
    return;
}
get_debug_type($items);

// Dynamic calls and names in strings.
$fn = 'array_is_list';
$fn([]);
call_user_func('array_is_list', []);
