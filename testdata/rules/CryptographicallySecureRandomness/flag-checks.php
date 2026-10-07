<?php
class Tokens
{
    public function truthy()
    {
        $bytes = openssl_random_pseudo_bytes(16, $strong);
        if ($bytes === false) {
            return null;
        }
        if ($strong) {
            return $bytes;
        }
        return null;
    }

    public function ternary()
    {
        $bytes = openssl_random_pseudo_bytes(16, $ok);
        return $bytes !== false && $ok ? $bytes : null;
    }

    public function loose()
    {
        $bytes = openssl_random_pseudo_bytes(16, $strong);
        if ($bytes === false || $strong == false) {
            return null;
        }
        return $bytes;
    }

    public function looseTrue()
    {
        $bytes = openssl_random_pseudo_bytes(16, $strong);
        if ($bytes === false || true != $strong) {
            return null;
        }
        return $bytes;
    }

    public function unchecked()
    {
        $bytes = openssl_random_pseudo_bytes(16, <error descr="The strength flag may be false; check it.">$strong</error>);
        if ($bytes === false || $strong == 1) {
            return null;
        }
        return [$bytes, $strong];
    }
}
