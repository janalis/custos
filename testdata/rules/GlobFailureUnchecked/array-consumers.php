<?php
function names($dir) {
    $paths = glob($dir . '/*.log');
    return array_map('basename', <warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function total($dir) { return count(<warning descr="Handle glob failure before using its result as an array.">glob($dir . '/*')</warning>); }
function joined($dir) {
    $paths = glob($dir . '/*.csv');
    return \implode(', ', <warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function known($dir, $name) {
    $paths = glob($dir . '/*');
    return in_array(needle: $name, haystack: <warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function guardedElsewhere($dir, $other) {
    $paths = glob($dir . '/*');
    if ($other) {
        return [];
    }
    foreach (<warning descr="Handle glob failure before iterating its result.">$paths</warning> as $path) {
        echo $path;
    }
    return array_values(<warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function nullOnly($dir) {
    $paths = glob($dir . '/*');
    if (isset($paths)) {
        foreach (<warning descr="Handle glob failure before iterating its result.">$paths</warning> as $path) {
            echo $path;
        }
    }
}
function reassigned($dir) {
    $paths = glob($dir . '/a*');
    if (!$paths) {
        return;
    }
    $paths = glob($dir . '/b*');
    foreach (<warning descr="Handle glob failure before iterating its result.">$paths</warning> as $path) {
        echo $path;
    }
}
