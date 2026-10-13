<?php
// The signature itself proves success before it is published.
function signEmptyGuard(string $data, $key): ?string
{
    openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256);
    if (empty($signature)) {
        return null;
    }
    return base64_encode($signature);
}
function signIssetTernary(string $data, $key): ?string
{
    openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256);
    return isset($signature) ? bin2hex($signature) : null;
}
function signVerified(string $data, $key, $public): string
{
    openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256);
    if (openssl_verify($data, (string) $signature, $public, OPENSSL_ALGO_SHA256) !== 1) {
        throw new RuntimeException('signing failed');
    }
    return base64_encode($signature);
}
function signInitialised(array $items, $key): array
{
    $out = [];
    foreach ($items as $data) {
        $signature = '';
        openssl_sign($data, $signature, $key);
        if (!$signature) {
            continue;
        }
        $out[] = base64_encode($signature);
    }
    return $out;
}
function signBranches(string $data, $key): void
{
    $signature = null;
    openssl_sign($data, $signature, $key);
    if (isset($signature) && $data !== '') {
        echo base64_encode($signature);
    }
    if (empty($signature)) {
        echo 'failed';
    } else {
        echo bin2hex($signature);
    }
}
