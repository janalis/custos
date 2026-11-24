<?php
// The variable holds a closure that draws a new nonce on each call.
function seal(string $key)
{
    $n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES, ...);
    sodium_crypto_secretbox('a', $n, $key);
    sodium_crypto_secretbox('b', $n, $key);
}
