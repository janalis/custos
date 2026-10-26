<?php
const OPTION_NAMES = [
    \CURLOPT_PROXY => 'proxy',
    \CURLOPT_SSL_VERIFYPEER => 'verify_peer',
    \CURLOPT_SSL_VERIFYHOST => "verify_host",
];
curl_setopt_array($h, [\CURLOPT_USERAGENT => 'agent', <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">\CURLOPT_SSL_VERIFYHOST => '0'</error>]);
curl_setopt_array($h, [\CURLOPT_USERAGENT => 'agent', <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">\CURLOPT_SSL_VERIFYPEER => ' '</error>]);
curl_setopt_array($h, [\CURLOPT_USERAGENT => 'agent', <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">\CURLOPT_SSL_VERIFYPEER => 'no'</error>, \CURLOPT_TIMEOUT => 5]);
curl_setopt_array($h, [\CURLOPT_USERAGENT, <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">\CURLOPT_SSL_VERIFYPEER => 'no'</error>]);
