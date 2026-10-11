<?php
function sealBlocks($data, $key, $cipher)
{
    // ECB takes no IV: the empty string is the only correct value.
    $a = openssl_encrypt($data, 'aes-128-ecb', $key, OPENSSL_RAW_DATA, '');
    $b = openssl_encrypt($data, ('AES-256-ECB'), $key, 0, "");
    $c = mcrypt_encrypt('rijndael-128', $key, $data, 'ecb', '');
    $d = mcrypt_encrypt('rijndael-128', $key, $data, MCRYPT_MODE_ECB, '');
    $e = openssl_encrypt(iv: '', data: $data, cipher_algo: 'aes-128-ecb', passphrase: $key);

    $f = openssl_encrypt($data, 'aes-128-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: ''.">''</error>);
    $g = openssl_encrypt($data, $cipher, $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: ''.">''</error>);
    $h = mcrypt_encrypt('rijndael-128', $key, $data, MCRYPT_MODE_CBC, <error descr="Generate the IV with mcrypt_create_iv(); it may come from: ''.">''</error>);
    $i = openssl_encrypt($data, 'aes-128-ecb', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'abc'.">'abc'</error>);
    $j = openssl_encrypt($data, 'aes-128-ecb', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 0.">0</error>);
    $k = openssl_encrypt($data, 'aes-128-ecb', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: mt_rand().">mt_rand()</error>);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
