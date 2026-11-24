<?php
// A partial application of random_bytes() is a closure, not a 16-byte key.
function seal($m, $n)
{
    $k = random_bytes(16, ...);
    return sodium_crypto_secretbox($m, $n, $k);
}
