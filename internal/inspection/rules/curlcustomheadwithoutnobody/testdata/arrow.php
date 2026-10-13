<?php

function capturedByArrowFunction($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $quiet = fn () => curl_setopt($h, CURLOPT_NOBODY, true);
    $quiet();
    return curl_exec($h);
}

function unrelatedArrowFunction($url, array $rows)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $ids = array_map(fn ($row) => $row['id'], $rows);
    return <warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
}
