<?php
function fetch($handle, $insecure, $opts)
{
    curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, $insecure ? 0 : 2);
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, '1');
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, true);
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, $insecure ? 0 : true);
    curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, getenv('VERIFY'));
    curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, 10);
    curl_setopt($handle, CURLOPT_SSL_VERIFYHOST);
    curl_setopt($handle, CURLOPT_TIMEOUT, 0);
    $list = [CURLOPT_SSL_VERIFYPEER];
    $opts[CURLOPT_SSL_VERIFYPEER] ??= 0;
    $set = isset($opts[CURLOPT_SSL_VERIFYHOST]);
    $same = CURLOPT_SSL_VERIFYHOST == 0;
    return [$list, $set, $same];
}
curl_setopt($h, CURLOPT_SSL_VERIFYPEER, $top);
