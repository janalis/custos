<?php

function testedBeforeUse($url)
{
    $h = curl_init($url);
    if ($h === false || !$h) {
        return null;
    }
    if (($h)) {
        log_ready();
    } elseif ($h) {
        return null;
    }
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    if (empty($h) && isset($h)) {
        return null;
    }
    return <warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
}

function unrelatedClosure($url, array $rows)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $prefix = 'id-';
    usort($rows, function ($a, $b) use ($prefix) {
        return strcmp($prefix . $a, $prefix . $b);
    });
    $later = static function () {
        $h = [$h ?? null];
        return $h;
    };
    return <warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
}

function escapesOnlyAfterwards($url)
{
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $result = <warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
    return [$h, $result];
}
