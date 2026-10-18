<?php
function storedInArray() {
    $handle = curl_init();
    $saved = [$handle];
    curl_setopt($handle, CURLOPT_HTTPHEADER, ['Authorization: Bearer example']);
    curl_setopt($handle, CURLOPT_HTTPHEADER, ['Accept: application/json']);
    curl_exec($handle);
}
function capturedHandle() {
    $handle = curl_init();
    $later = fn () => $handle;
    curl_setopt($handle, CURLOPT_HTTPHEADER, ['Authorization: Bearer example']);
    curl_setopt($handle, CURLOPT_HTTPHEADER, ['Accept: application/json']);
    curl_exec($handle);
}
