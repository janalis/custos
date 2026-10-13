<?php
define('VERIFY_HOST', '2');
define('VERIFY_OFF', 0);
define('PEER_ON', true);
define('PEER_NULL', null);
define('PEER_ALIAS', UNKNOWN_SETTING);

function more($h, $v, $options)
{
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($h, CURLOPT_SSL_VERIFYHOST, "{$v}")</error>;
    curl_setopt($h, CURLOPT_SSL_VERIFYHOST, VERIFY_HOST);
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($h, CURLOPT_SSL_VERIFYHOST, VERIFY_OFF)</error>;
    curl_setopt($h, CURLOPT_SSL_VERIFYPEER, PEER_ON);
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($h, CURLOPT_SSL_VERIFYPEER, PEER_NULL)</error>;
    curl_setopt($h, CURLOPT_SSL_VERIFYPEER, UNDEFINED_CONST);
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($h, CURLOPT_SSL_VERIFYPEER, PEER_ALIAS)</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">$options[CURLOPT_SSL_VERIFYHOST]['x'] = 0</error>;
    $first = CURLOPT_SSL_VERIFYHOST[0];
    $peer = $peer2 = $peer3 = false;
    $host = 0;
    list(CURLOPT_SSL_VERIFYPEER => $peer) = $options;
    [CURLOPT_SSL_VERIFYPEER => $peer2] = $options;
    ['a' => [CURLOPT_SSL_VERIFYPEER => $peer3]] = $options;
    foreach ($options as [CURLOPT_SSL_VERIFYHOST => $host]) {}
    curl_setopt($h, CURLOPT_SSL_VERIFYPEER, ...$v);
    $partial = curl_setopt($h, CURLOPT_SSL_VERIFYPEER, ...); // not valid PHP: placeholder only alone
    $copy = [<error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">CURLOPT_SSL_VERIFYPEER => $peer</error>];
    list('a' => [CURLOPT_SSL_VERIFYPEER => $peer]) = $options;
    $nested = [[<error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">CURLOPT_SSL_VERIFYHOST => 0</error>]];
    $x = [[CURLOPT_SSL_VERIFYHOST => $h2]] = $options;
    foreach ($options as $k => $o) { $o = [CURLOPT_SSL_VERIFYHOST => 2]; }
    return [$first, $peer, $peer2, $peer3, $host];
}
