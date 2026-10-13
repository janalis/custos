<?php
class Box
{
    private $fresh;

    public function lock($msg, $pass)
    {
        $this->fresh = random_bytes(16);
        $ok1 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, $this->fresh);
        $ok2 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, $this->makeIv());
        $ok3 = openssl_encrypt($msg, 'aes-256-cbc', $pass);
        $ok5 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, self::makeStaticIv());
        $ok6 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, makeIv());
        $ok7 = mcrypt_encrypt('rijndael-128', $pass, $msg, 'cbc', mcrypt_create_iv(16));
        return [$ok1, $ok2, $ok3, $ok5, $ok6, $ok7];
    }

    private function makeIv()
    {
        return openssl_random_pseudo_bytes(16);
    }

    private static function makeStaticIv()
    {
        $bytes = random_bytes(16);
        return $bytes;
    }
}

function makeIv()
{
    return \random_bytes(16);
}

openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, $topLevel);
