<?php
function secureOption()
{
    $request = curl_init();
    curl_setopt($request, CURLOPT_URL, "https://intranet.example.test/report");
    curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2");
    curl_exec($request);
}

function upgradedLater()
{
    $request = curl_init("http://intranet.example.test/report");
    curl_setopt($request, CURLOPT_URL, "https://intranet.example.test/report");
    curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2");
    curl_exec($request);
}

function unknownTarget($target)
{
    $request = curl_init();
    curl_setopt($request, CURLOPT_URL, $target);
    curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2");
    curl_exec($request);
}

function noTarget()
{
    $request = curl_init();
    curl_setopt($request, CURLOPT_USERPWD, "reader:hunter2");
    curl_exec($request);
}
