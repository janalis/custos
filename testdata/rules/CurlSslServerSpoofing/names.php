<?php
namespace Http;

function curl_setopt($h, $o, $v) {}

function fetch($handle)
{
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, false);
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">\Curl_SetOpt($handle, CURLOPT_SSL_VERIFYPEER, false)</error>;
}
