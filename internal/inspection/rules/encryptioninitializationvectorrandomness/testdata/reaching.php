<?php
function seal(string $msg, string $key, bool $legacy): array
{
    $iv = random_bytes(16);
    $box = openssl_encrypt($msg, 'aes-256-cbc', $key, 0, $iv);
    $iv = base64_encode($iv); // after the call: not the IV used above

    $nonce = openssl_random_pseudo_bytes(16);
    if ($legacy) {
        $nonce = md5($key);
    }
    $old = openssl_encrypt($msg, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: md5($key).">$nonce</error>);
    return [$box, $iv, $old];
}
