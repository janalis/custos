<?php
class Vault
{
    public function seal($msg, $pass, $m)
    {
        $a = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, ...);
        $b = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $this->$m().">$this->$m()</error>);
        $c = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: self::$m().">self::$m()</error>);
        $d = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $this->viaMethods().">$this->viaMethods()</error>);
        return [$a, $b, $c, $d];
    }

    private function viaMethods()
    {
        $this->prepare();
        $this->$name();
        return self::pad(Vault::$m());
    }

    private function prepare() {}
    private static function pad($s) { return $s; }
}
