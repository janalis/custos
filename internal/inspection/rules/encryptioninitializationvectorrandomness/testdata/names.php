<?php
namespace Vault;

function openssl_encrypt($d, $m, $k, $o, $iv) { return $d; }
function random_bytes($n) { return str_repeat('0', $n); }

class Safe
{
    public function own($msg, $pass)
    {
        return openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, 'fixed-iv');
    }

    public function builtin($msg, $pass)
    {
        $secure = \OpenSSL_Encrypt($msg, 'aes-256-cbc', $pass, 0, \Random_Bytes(16));
        $fake = random_bytes(16);
        return \OPENSSL_ENCRYPT($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: random_bytes(16).">$fake</error>);
    }
}
