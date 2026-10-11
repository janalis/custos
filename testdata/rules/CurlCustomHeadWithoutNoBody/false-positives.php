<?php

function configureAll(array $context)
{
    curl_setopt($context['handle'], CURLOPT_NOBODY, true);
}

function capturedByClosure($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $quiet = function () use ($h) {
        curl_setopt($h, CURLOPT_NOBODY, true);
    };
    $quiet();
    return curl_exec($h);
}

function passedInsideArray($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    configureAll(['handle' => $h]);
    return curl_exec($h);
}

function copyPassedInsideArray($url)
{
    $h = curl_init($url);
    $copy = $h;
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    configureAll(array('handle' => $copy));
    return curl_exec($h);
}

function storedInElement($url, array &$pool)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $pool['head'] = $h;
    prepare($pool);
    return curl_exec($h);
}

function boundByReference($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $other = &$h;
    return curl_exec($h);
}

function dynamicVariable($url, $name)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    prepare($$name);
    return curl_exec($h);
}

function suppressedRead($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $kept = @$h;
    return curl_exec($h);
}

function fallbackRead($url, $spare)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $kept = $h ?: $spare;
    return curl_exec($h);
}

function passedByName($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    configureAll(compact('h'));
    return curl_exec($h);
}
