<?php
namespace Crypto;

function openssl_random_pseudo_bytes($n) { return str_repeat('x', $n); }

class Keys
{
    public function own()
    {
        return openssl_random_pseudo_bytes(16);
    }

    public function builtin()
    {
        $raw = \<error descr="Pass a second argument to learn whether a strong algorithm was used.">OpenSSL_Random_Pseudo_Bytes</error>(16);
        if ($raw === false) {
            return null;
        }
        return $raw;
    }
}
