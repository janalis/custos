<?php
// Other scopes that never see the handle do not weaken the proof.
function unrelatedCapture($url) {
    $ch = curl_init($url);
    $retries = 3;
    $count = function () use ($retries) { return $retries; };
    $double = fn () => $retries * 2;
    return <warning descr="Enable response-body return before decoding curl output.">json_decode(curl_exec($ch))</warning>;
}
function shadowedInsideClosure($url) {
    $ch = curl_init($url);
    $local = function () { $ch = 1; return $ch; };
    return <warning descr="Enable response-body return before decoding curl output.">json_decode(curl_exec($ch))</warning>;
}
function handleCreatedInsideClosure($url) {
    return function () use ($url) {
        $ch = curl_init($url);
        return <warning descr="Enable response-body return before decoding curl output.">json_decode(curl_exec($ch))</warning>;
    };
}
function handleCreatedInsideArrowFunction($url) {
    return fn () => <warning descr="Enable response-body return before decoding curl output.">json_decode(curl_exec(curl_init($url)))</warning>;
}
function unrelatedGlobal() {
    global $settings;
    return $settings;
}
$plain = curl_init('https://example.invalid/');
<warning descr="Enable response-body return before decoding curl output.">json_decode(curl_exec($plain))</warning>;
