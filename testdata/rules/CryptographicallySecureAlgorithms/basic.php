<?php
namespace Vault;

final class Sealer
{
    public function seal($key, $data)
    {
        $legacy = mcrypt_encrypt(<error descr="Weak algorithm selected via MCRYPT_TRIPLEDES (3DES); prefer MCRYPT_RIJNDAEL_128.">MCRYPT_TRIPLEDES</error>, $key, $data, 'cbc');
        $wide   = [<error descr="Weak algorithm selected via MCRYPT_RIJNDAEL_256 (Rijndael with a 256-bit block is not AES); prefer MCRYPT_RIJNDAEL_128 with a 256-bit key.">\MCRYPT_RIJNDAEL_256</error>];
        if (<error descr="Weak algorithm selected via CRYPT_MD5 (MD5); prefer CRYPT_BLOWFISH.">CRYPT_MD5</error> === 1) {
            return openssl_encrypt($data, 'aes-128-gcm', $key);
        }
        $c = <error descr="Weak algorithm selected via OPENSSL_CIPHER_RC2_64 (RC2); prefer an AES-128-* cipher.">OPENSSL_CIPHER_RC2_64</error>;
        return [$legacy, $wide, $c, MCRYPT_RIJNDAEL_128, CRYPT_BLOWFISH, OPENSSL_ALGO_SHA1];
    }
}
