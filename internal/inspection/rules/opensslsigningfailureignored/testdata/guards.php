<?php
// A failed call leaves the variable untouched: stale content defeats the test.
function signAll(array $items, array $keys): array
{
    $out = [];
    foreach ($items as $i => $data) {
        <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $keys[$i], OPENSSL_ALGO_SHA256)</warning>;
        if (empty($signature)) {
            continue;
        }
        $out[] = base64_encode($signature);
    }
    return $out;
}
function signParameter(string $data, $signature, $key): ?string
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    return isset($signature) ? base64_encode($signature) : null;
}
function signInitialisedEmpty(string $data, $key): ?string
{
    $signature = '';
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    return isset($signature) ? base64_encode($signature) : null;
}
function signPublishedFirst(string $data, $key): void
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    echo base64_encode($signature);
    if (empty($signature)) {
        return;
    }
}
function signWrongBranch(string $data, $key): void
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    if (empty($signature)) {
        echo base64_encode($signature);
    }
}
// Negating or testing a value computed from the signature publishes it first.
function signSentNegated(string $data, $key): bool
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    if (!deliver(base64_encode($signature))) {
        return false;
    }
    return true;
}
function signSentEmpty(string $data, $key): bool
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    return empty(deliver(base64_encode($signature)));
}
function signLookedUp(string $data, array $seen, $key): bool
{
    <warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key)</warning>;
    return isset($seen[base64_encode($signature)]);
}
