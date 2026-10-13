<?php
class CipherMatrixTest
{
    public function provider()
    {
        return [MCRYPT_DES, OPENSSL_CIPHER_RC2_40];
    }
}
function other() {
    return [mcrypt_des, Cipher::MCRYPT_DES, 'des-ede3', md5('x'), OPENSSL_ALGO_MD5];
}
