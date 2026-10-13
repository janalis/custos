<?php
function fetch($handle, $insecure)
{
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, 1)</error>;
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, FALSE)</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, $insecure ? 0 : '1')</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($handle, \CURLOPT_SSL_VERIFYHOST, true)</error>;

    $peer = null;
    curl_setopt_array($handle, [
        CURLOPT_TIMEOUT => 5,
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">CURLOPT_SSL_VERIFYPEER => $peer</error>,
        CURLOPT_SSL_VERIFYHOST => "2",
    ]);

    $conf = [];
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">$conf['tls'][CURLOPT_SSL_VERIFYHOST] = true</error>;
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">$conf[CURLOPT_SSL_VERIFYPEER] = '0'</error>;
    return $conf;
}
