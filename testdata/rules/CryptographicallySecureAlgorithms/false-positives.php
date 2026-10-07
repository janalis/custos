<?php
namespace App\Tests\Unit;

class CipherMatrix
{
    public function provider()
    {
        return [MCRYPT_DES, OPENSSL_CIPHER_RC2_40];
    }
}
