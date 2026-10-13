<?php
function encrypt(string $data, string $key, IvSource $src)
{
    $a = openssl_encrypt($data, 'aes-256-cbc', $key, 0, make_iv(16));
    $b = openssl_encrypt($data, 'aes-256-cbc', $key, 0, $src->iv());
    $c = openssl_encrypt($data, 'aes-256-cbc', $key, 0, IvSource::make());
    $d = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: weak_iv().">weak_iv()</error>);
    $e = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $src->weak().">$src->weak()</error>);
    return [$a, $b, $c, $d, $e];
}
