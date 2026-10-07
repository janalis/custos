<?php
function lib($h)
{
    curl_setopt($h, CURLOPT_SSL_VERIFYHOST, LIB_HOST);
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($h, CURLOPT_SSL_VERIFYHOST, LIB_HOST_BAD)</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($h, CURLOPT_SSL_VERIFYHOST, LIB_PEER)</error>;
    curl_setopt($h, CURLOPT_SSL_VERIFYPEER, LIB_PEER);
    curl_setopt($h, CURLOPT_SSL_VERIFYPEER, LIB_ONE);
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($h, CURLOPT_SSL_VERIFYPEER, LIB_PEER_OFF)</error>;
    curl_setopt($h, CURLOPT_SSL_VERIFYHOST, LIB_TEN);
}
