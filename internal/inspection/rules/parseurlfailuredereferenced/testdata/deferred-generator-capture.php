<?php
function unstarted_generator(string $path) {
    $parts = parse_url($path);
    if ($parts === false) { return; }
    $replace = function () use (&$parts, $path) { $parts = parse_url($path); yield 1; };
    $replace();
    echo $parts['host'];
}
function immediate_capture(string $path) {
    $parts = parse_url($path);
    if ($parts === false) { return; }
    $replace = function () use (&$parts, $path) { $parts = parse_url($path); };
    $replace();
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
function syntactic_generator(string $path) {
    $parts = parse_url($path);
    if ($parts === false) { return; }
    $replace = function () use (&$parts, $path) {
        $parts = parse_url($path);
        if (false) { yield 1; }
    };
    $replace();
    echo $parts['host'];
}
function nested_generator_is_independent(string $path) {
    $parts = parse_url($path);
    if ($parts === false) { return; }
    $replace = function () use (&$parts, $path) {
        $parts = parse_url($path);
        $nested = function () { yield 1; };
    };
    $replace();
    echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}
