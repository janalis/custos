<?php

function fetchAfterSinkReset($url, $fp)
{
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_FILE, $fp);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_TIMEOUT, 10);
    if (!<warning descr="Compare cURL transport failure strictly with false.">curl_exec($ch)</warning>) {
        throw new RuntimeException(curl_error($ch));
    }
}

function fetchWithOptionList($url)
{
    $ch = curl_init($url);
    curl_setopt_array($ch, [CURLOPT_FOLLOWLOCATION => true, CURLOPT_RETURNTRANSFER => true]);
    if (!!!<warning descr="Compare cURL transport failure strictly with false.">curl_exec($ch)</warning>) {
        return false;
    }
    return true;
}
