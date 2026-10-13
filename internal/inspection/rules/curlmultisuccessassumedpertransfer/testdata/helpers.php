<?php

// Lifecycle builtins and helpers that never receive the handle keep the report.
function fetchUnchecked($url, $logger)
{
    $mh = curl_multi_init();
    curl_multi_add_handle($mh, curl_init($url));
    curl_multi_setopt($mh, CURLMOPT_MAXCONNECTS, 4);
    $logger->info($url);
    do {
        $status = <warning descr="Inspect each completed transfer result.">curl_multi_exec($mh, $running)</warning>;
        if ($running) {
            curl_multi_select($mh);
        }
    } while ($running && $status === CURLM_OK);
    if ($status === CURLM_OK && !$running) {
        return true;
    }
    return false;
}
