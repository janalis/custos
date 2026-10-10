<?php
$key=random_bytes(SODIUM_CRYPTO_SECRETBOX_KEYBYTES); $nonce=random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); $a=sodium_crypto_secretbox('a',$nonce,$key); sodium_increment($nonce); $b=sodium_crypto_secretbox('b',$nonce,$key); echo strlen($a).' '.strlen($b);

function rotate_key($nonce, $key) {
    sodium_crypto_secretbox('first', $nonce, $key);
    sodium_increment(string: $key);
    sodium_crypto_secretbox('second', $nonce, $key);
}
function add_nonce($nonce, $key, $step) {
    sodium_crypto_secretbox('first', $nonce, $key);
    sodium_add(string1: $nonce, string2: $step);
    sodium_crypto_secretbox('second', $nonce, $key);
}
function unrelated_nonce($nonce, $key, $other) {
    sodium_crypto_secretbox('first', $nonce, $key);
    sodium_increment($other);
    <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('second', $nonce, $key)</warning>;
}

function reuse_after_increment($nonce, $key) {
    sodium_crypto_secretbox('first', $nonce, $key);
    sodium_increment($nonce);
    sodium_crypto_secretbox('second', $nonce, $key);
    <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('third', $nonce, $key)</warning>;
}
