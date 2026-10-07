<?php
function seal($msg, $key, $aad)
{
    $a = openssl_encrypt($msg, 'aes-256-gcm', $key, OPENSSL_RAW_DATA, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'nonce-000001'.">'nonce-000001'</error>, $tag);
    $b = openssl_encrypt($msg, 'aes-256-gcm', $key, OPENSSL_RAW_DATA, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: uniqid().">uniqid()</error>, $tag, $aad, 16);
    $c = openssl_encrypt($msg, 'aes-256-gcm', $key, OPENSSL_RAW_DATA, random_bytes(12), $tag, $aad);
    return [$a, $b, $c, $tag];
}
