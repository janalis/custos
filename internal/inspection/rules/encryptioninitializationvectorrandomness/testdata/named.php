<?php
function sealNamed(string $data, string $key, array $rest): array
{
    $tag = null;
    $a = openssl_encrypt($data, 'aes-256-gcm', $key, OPENSSL_RAW_DATA, tag: $tag, iv: random_bytes(12));
    $b = openssl_encrypt($data, 'aes-256-cbc', $key, iv: <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: str_repeat(&quot;\0&quot;, 16).">str_repeat("\0", 16)</error>);
    $c = openssl_encrypt($data, 'aes-256-gcm', $key, 0, ...$rest);
    $d = openssl_encrypt(iv: <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'fixed'.">'fixed'</error>, data: $data, cipher_algo: 'aes-256-cbc', passphrase: $key);
    return [$a, $b, $c, $d, $tag];
}
