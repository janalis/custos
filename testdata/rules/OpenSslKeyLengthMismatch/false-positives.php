<?php
// A by-reference capture lets the closure replace the placeholder.
function encryptWithLoader(string $data, string $iv, callable $loader): string
{
    $key = 'unset';
    $load = function () use (&$key, $loader): void {
        $key = $loader();
    };
    $load();
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

// A global binding is visible to every later call.
function encryptWithGlobal(string $data, string $iv): string
{
    global $appKey;
    $appKey = 'unset';
    load_app_key();
    return openssl_encrypt($data, 'aes-256-cbc', $appKey, 0, $iv);
}

// A reference stored in an array escapes with the array.
function encryptWithSlots(string $data, string $iv): string
{
    $key = 'unset';
    $slots = [&$key];
    fill_slots($slots);
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

// File-level variables are globals too.
$fileKey = 'unset';
load_file_key();
openssl_encrypt($payload, 'aes-256-cbc', $fileKey, 0, $nonce);
