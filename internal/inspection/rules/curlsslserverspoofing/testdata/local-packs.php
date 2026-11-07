<?php
function localTlsPacks($handle) {
    $options = [CURLOPT_SSL_VERIFYPEER, false];
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($handle, ...$options)</error>;
    $safe = [CURLOPT_SSL_VERIFYPEER, true];
    curl_setopt($handle, ...$safe);
    $changed = [CURLOPT_SSL_VERIFYPEER, false];
    $changed[1] = true;
    curl_setopt($handle, ...$changed);
    $escaped = [CURLOPT_SSL_VERIFYPEER, false];
    unknown($escaped);
    curl_setopt($handle, ...$escaped);
}
