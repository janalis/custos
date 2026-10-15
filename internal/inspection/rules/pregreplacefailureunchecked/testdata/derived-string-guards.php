<?php
function derived_guards($source, $conditional) {
    $value = $source;
    if ($conditional) { $value = preg_replace('/x/', '', $value); }
    if (false === <warning descr="Handle regex replacement failure before using its result.">strpos($value, '$')</warning>) { return; }
    substr($value, 1);
}
function suffix_guard($source, $conditional) {
    $value = $source;
    if ($conditional) { $value = preg_replace('/x/', '', $value); }
    if (<warning descr="Handle regex replacement failure before using its result.">str_ends_with($value, '/')</warning>) { substr($value, 1); }
}
function offset_guard($source, $conditional) {
    $value = $source;
    if ($conditional) { $value = preg_replace('/x/', '', $value); }
    if ($value[0] === 'x') { substr($value, 1); }
}
function empty_suffix($source, $conditional) {
    $value = $source;
    if ($conditional) { $value = preg_replace('/x/', '', $value); }
    if (<warning descr="Handle regex replacement failure before using its result.">str_ends_with($value, '')</warning>) {
        <warning descr="Handle regex replacement failure before using its result.">substr($value, 1)</warning>;
    }
}
function changed_after_test($source, $conditional) {
    $value = $source;
    if ($conditional) { $value = preg_replace('/x/', '', $value); }
    if (<warning descr="Handle regex replacement failure before using its result.">str_ends_with($value, '/')</warning>) {
        $value = preg_replace('/y/', '', $value);
        <warning descr="Handle regex replacement failure before using its result.">substr($value, 1)</warning>;
    }
}
