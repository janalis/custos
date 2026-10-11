<?php
function fetchReport()
{
    $request = curl_init();
    curl_setopt($request, CURLOPT_URL, "http://intranet.example.test/report");
    <warning descr="Protect basic authentication with HTTPS.">curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2")</warning>;
    curl_exec($request);
}

function fetchOverridden($fallback)
{
    $request = curl_init($fallback);
    curl_setopt($request, CURLOPT_URL, "http://intranet.example.test/report");
    <warning descr="Protect basic authentication with HTTPS.">curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2")</warning>;
    curl_exec($request);
}
