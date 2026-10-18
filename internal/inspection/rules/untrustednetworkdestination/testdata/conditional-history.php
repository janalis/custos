<?php
function conditionalUrl($flag) {
    $handle = curl_init('https://api.example.com/');
    if ($flag) { curl_setopt($handle, CURLOPT_URL, $_GET['destination']); }
    <error descr="Validate the outbound network destination and redirects.">curl_exec($handle)</error>;
}
function conditionalUrlArray($flag) {
    $handle = curl_init('https://api.example.com/');
    if ($flag) { curl_setopt_array($handle, [CURLOPT_URL => $_GET['destination']]); }
    <error descr="Validate the outbound network destination and redirects.">curl_exec($handle)</error>;
}
function escapedUrl() {
    $handle = curl_init($_GET['destination']);
    $callback = fn () => curl_setopt($handle, CURLOPT_URL, 'https://api.example.com/');
    curl_exec($handle);
}
