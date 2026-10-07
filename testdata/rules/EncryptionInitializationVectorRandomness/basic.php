<?php
class Box
{
    const SALT = 'static-iv';
    private $nonce = 'fixed';
    private $fresh;

    public function lock($msg, $pass, $iv = '0000')
    {
        $iv = uniqid();
        $out = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: uniqid().">$iv</error>);
        $old = mcrypt_encrypt('rijndael-128', $pass, $msg, 'cbc', <error descr="Generate the IV with mcrypt_create_iv(); it may come from: 'static-iv'.">self::SALT</error>);

        $this->nonce = rand();
        $alt = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'fixed', rand().">$this->nonce</error>);
        $raw = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $this->rng()->random_bytes(16).">$this->rng()->random_bytes(16)</error>);
        $wrap = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $this->weakIv().">$this->weakIv()</error>);

        return [$out, $old, $alt, $raw, $wrap];
    }

    private function weakIv()
    {
        return str_repeat('a', 16);
    }
}
