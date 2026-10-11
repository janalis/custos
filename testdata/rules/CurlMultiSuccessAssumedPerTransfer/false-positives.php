<?php

function failedTransfers($mh)
{
    $failed = 0;
    while ($info = curl_multi_info_read($mh)) {
        if ($info['result'] !== CURLE_OK) {
            $failed++;
        }
    }
    return $failed;
}

// The helper receives the multi handle and inspects every completed transfer.
function fetchAll($url)
{
    $mh = curl_multi_init();
    curl_multi_add_handle($mh, curl_init($url));
    do {
        $status = curl_multi_exec($mh, $running);
        if ($running) {
            curl_multi_select($mh);
        }
    } while ($running && $status === CURLM_OK);
    if (failedTransfers($mh) > 0) {
        return false;
    }
    if ($status === CURLM_OK && !$running) {
        return true;
    }
    return false;
}

class Batch
{
    private $handle;

    // A method may read the transfer results from the handle it is given.
    public function viaMethod($url)
    {
        $mh = curl_multi_init();
        curl_multi_add_handle($mh, curl_init($url));
        $this->drain($mh);
        if (curl_multi_exec($mh, $running) === CURLM_OK && !$running) {
            return true;
        }
        return false;
    }

    // A stored copy is reachable from code this function cannot see.
    public function viaProperty($url)
    {
        $mh = curl_multi_init();
        curl_multi_add_handle($mh, curl_init($url));
        $this->handle = $mh;
        $this->drain(null);
        if (curl_multi_exec($mh, $running) === CURLM_OK && !$running) {
            return true;
        }
        return false;
    }

    private function drain($mh)
    {
    }
}
