<?php
// The nonce or the key is rebound or altered between the two encryptions.
function short_destructuring(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    [$nonce] = [random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES)];
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function list_destructuring(string $key, array $pair): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    list(, $key) = $pair;
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function loop_value(string $key, array $nonces): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    foreach ($nonces as $nonce) {
    }
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function loop_key(string $key, array $byNonce): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    foreach ($byNonce as $nonce => $ignored) {
    }
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function offset_write(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    $nonce[0] = 'x';
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function counter_step(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    $nonce++;
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function reference_alias(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    $alias = &$nonce;
    $alias = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function extracted(string $key, array $fresh): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    extract($fresh);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function variable_variable(string $key, string $which): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    $$which = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function output_argument(string $key, string $pool): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    preg_match('/.{24}/s', $pool, $nonce);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function variadic_output(string $key, string $pool): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    sscanf($pool, '%s %s %s', $first, $second, $nonce);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function named_output(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    settype(var: $nonce, type: 'string');
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function refill(string &$bytes): void {
    $bytes = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
}
function user_output(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    refill($nonce);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
function fixed_nonce_new_key(array $keys): void {
    sodium_crypto_secretbox('invoice', 'fixed-nonce-24-bytes-long', $keys[0]);
    [$keys[0]] = [sodium_crypto_secretbox_keygen()];
    sodium_crypto_secretbox('receipt', 'fixed-nonce-24-bytes-long', $keys[0]);
}
// The write happens after the earlier encryption even when a later one
// repeats the same plaintext.
function same_plaintext_after_rebinding(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    [$nonce] = [random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES)];
    sodium_crypto_secretbox('receipt', $nonce, $key);
    sodium_crypto_secretbox('receipt', $nonce, $key);
}
