<?php

// A file sink set after RETURNTRANSFER makes curl_exec() return a boolean.
function downloadToFile($url, $fp)
{
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_FILE, $fp);
    if (!curl_exec($ch)) {
        throw new RuntimeException(curl_error($ch));
    }
}

// So does a write callback.
function streamBody($url)
{
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_WRITEFUNCTION, function ($ch, $chunk) {
        echo $chunk;
        return strlen($chunk);
    });
    if (!curl_exec($ch)) {
        throw new RuntimeException(curl_error($ch));
    }
}

// Options that cannot be read may switch RETURNTRANSFER off again.
function requestWithCallerOptions($url, array $extra, $option)
{
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt_array($ch, $extra);
    if (!curl_exec($ch)) {
        throw new RuntimeException();
    }

    $other = curl_init($url);
    curl_setopt($other, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($other, $option, false);
    if (!curl_exec($other)) {
        throw new RuntimeException();
    }

    $third = curl_init($url);
    curl_setopt_array($third, [CURLOPT_RETURNTRANSFER => true, CURLOPT_FILE => STDOUT]);
    if (!curl_exec($third)) {
        throw new RuntimeException();
    }
}

// An even number of negations guards the success path.
function rejectAnyBody($url)
{
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    if (!!curl_exec($ch)) {
        throw new RuntimeException('body received');
    }
    if (!(curl_exec($ch) == false)) {
        throw new RuntimeException('body received');
    }
}
