<?php
// A rebinding only clears the encryptions that precede it.
function reused_after_destructuring(string $key): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    [$nonce] = [random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES)];
    sodium_crypto_secretbox('receipt', $nonce, $key);
    <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('refund', $nonce, $key)</warning>;
}
function reused_inside_loop(string $key, array $nonces): void {
    foreach ($nonces as $nonce) {
        sodium_crypto_secretbox('invoice', $nonce, $key);
        <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('receipt', $nonce, $key)</warning>;
    }
}
// Writes that leave the nonce and the key alone do not hide the reuse.
function unrelated_writes(string $key, object $audit, array $slots): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    $audit->length = strlen(string: $nonce);
    $slots[$nonce] = bin2hex($nonce);
    [$first, $second] = $slots;
    $copy = &$first;
    $second++;
    foreach ($slots as $slot) {
    }
    preg_match('/^\d+$/', $nonce, $matches);
    $later = static function () use ($key) {
        $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
        return sodium_crypto_secretbox('archive', $nonce, $key);
    };
    <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('receipt', $nonce, $key)</warning>;
}
function surplus_arguments(string $key, string $pool): void {
    $nonce = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES);
    sodium_crypto_secretbox('invoice', $nonce, $key);
    preg_match('/x/', $pool, $matches, 0, 0, $nonce);
    <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('receipt', $nonce, $key)</warning>;
}
