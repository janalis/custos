<?php
namespace Legacy\Ciphers {
    const MCRYPT_DES = 'aes-256-gcm';
    const CRYPT_MD5 = 'argon2id';

    function pick() {
        return [MCRYPT_DES, namespace\CRYPT_MD5, <error descr="Weak algorithm selected via MCRYPT_RC4 (RC4); prefer MCRYPT_RIJNDAEL_128.">MCRYPT_RC4</error>];
    }
}

namespace Client {
    use const Legacy\Ciphers\CRYPT_MD5;

    function choose() {
        return [
            CRYPT_MD5,
            \Legacy\Ciphers\MCRYPT_DES,
            Other\OPENSSL_CIPHER_DES,
            <error descr="Weak algorithm selected via MCRYPT_DES (DES); prefer MCRYPT_RIJNDAEL_128.">\MCRYPT_DES</error>,
            <error descr="Weak algorithm selected via OPENSSL_CIPHER_DES (DES); prefer an AES-128-* cipher.">OPENSSL_CIPHER_DES</error>,
        ];
    }
}
