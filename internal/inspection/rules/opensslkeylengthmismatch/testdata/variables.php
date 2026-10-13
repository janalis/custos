<?php
function encryptWithShortKey(string $data, string $iv): string
{
    $key = 'too-short';
    audit($data);
    return openssl_encrypt($data, 'aes-256-cbc', <warning descr="Provide the cipher-required key length.">$key</warning>, 0, $iv);
}

function encryptWithValueCapture(string $data, string $iv): string
{
    $key = str_repeat('k', 16);
    $describe = function () use ($key): int {
        return strlen($key);
    };
    $describe();
    return openssl_encrypt($data, 'aes-192-cbc', <warning descr="Provide the cipher-required key length.">$key</warning>, 0, $iv);
}

// Builtins without a callback parameter cannot replace a file-level variable.
$scriptKey = 'too-short';
$scriptIv = random_bytes(16);
echo base64_encode(openssl_encrypt($payload, 'aes-256-cbc', <warning descr="Provide the cipher-required key length.">$scriptKey</warning>, 0, $scriptIv));
