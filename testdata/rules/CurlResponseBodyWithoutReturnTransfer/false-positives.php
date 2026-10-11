<?php
// The handle is shared with another variable scope that may configure it.
function capturedByClosure($url) {
    $ch = curl_init($url);
    $configure = function () use ($ch) { curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); };
    $configure();
    return json_decode(curl_exec($ch));
}
function aliasCapturedByClosure($url) {
    $ch = curl_init($url);
    $same = $ch;
    $configure = function () use ($same) { curl_setopt($same, CURLOPT_RETURNTRANSFER, true); };
    $configure();
    return json_decode(curl_exec($ch));
}
function referencedByArrowFunction($url) {
    $ch = curl_init($url);
    $configure = fn () => curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $configure();
    return json_decode(curl_exec($ch));
}
function configureShared() {
    global $shared;
    curl_setopt($shared, CURLOPT_RETURNTRANSFER, true);
}
$shared = curl_init('https://example.invalid/');
configureShared();
json_decode(curl_exec($shared));
