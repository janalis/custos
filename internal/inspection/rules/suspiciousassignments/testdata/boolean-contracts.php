<?php
function measured() {
    $asNumber = false;
    [$seconds, $nanoseconds] = hrtime($asNumber);
    return $seconds + $nanoseconds;
}
function dynamicRead($value) {
    $getter = fn() => $value;
    $value = 1;
    return $value + $getter();
}
function jumpedRead($value) {
    goto observe;
    $value = 1;
observe:
    return $value;
}
function branchRead($value, $condition) {
    if ($condition) {
        return $value;
    }
    $value = 1;
    return $value;
}
function tableRead($value) {
    $saved = get_defined_vars();
    $value = 1;
    return [$saved, $value];
}
function indirectRead($value, $name) {
    $saved = $$name;
    $value = 1;
    return [$saved, $value];
}
