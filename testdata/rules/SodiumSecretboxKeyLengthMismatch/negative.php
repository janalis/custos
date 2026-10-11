<?php
sodium_crypto_secretbox($message, $nonce, sodium_crypto_secretbox_keygen());

function referenceAlias($m, $n)
{
    $k = random_bytes(16);
    $r = &$k;
    $r = random_bytes(32);
    return sodium_crypto_secretbox($m, $n, $k);
}

function referenceInArray($m, $n)
{
    $k = random_bytes(16);
    foreach ([&$k] as &$slot) {
        $slot = random_bytes(32);
    }
    return sodium_crypto_secretbox($m, $n, $k);
}

function closureByReference($m, $n)
{
    $k = random_bytes(16);
    $f = function () use (&$k) {
        $k = random_bytes(32);
    };
    $f();
    return sodium_crypto_secretbox($m, $n, $k);
}

function extracted($m, $n)
{
    $k = random_bytes(16);
    extract(['k' => random_bytes(32)]);
    return sodium_crypto_secretbox($m, $n, $k);
}

function variableVariable($m, $n)
{
    $k = random_bytes(16);
    $name = 'k';
    $$name = random_bytes(32);
    return sodium_crypto_secretbox($m, $n, $k);
}

function shortDestructuring($m, $n)
{
    $k = random_bytes(16);
    [$k] = [random_bytes(32)];
    return sodium_crypto_secretbox($m, $n, $k);
}

function nestedListDestructuring($m, $n)
{
    $k = random_bytes(16);
    list('a' => list($k)) = ['a' => [random_bytes(32)]];
    return sodium_crypto_secretbox_open($m, $n, $k);
}

function foreachTarget($m, $n, $keys)
{
    $k = random_bytes(16);
    foreach ($keys as $k) {
    }
    return sodium_crypto_secretbox($m, $n, $k);
}

function foreachDestructuring($m, $n, $rows)
{
    $k = random_bytes(16);
    foreach ($rows as [$k, $label]) {
    }
    return sodium_crypto_secretbox($m, $n, $k);
}

function globalImport($m, $n)
{
    $k = random_bytes(16);
    global $k;
    return sodium_crypto_secretbox($m, $n, $k);
}

function builtinOutParameter($m, $n, $text)
{
    $k = random_bytes(16);
    preg_match('/.{32}/', $text, $k);
    return sodium_crypto_secretbox($m, $n, $k);
}

function methodArgument($m, $n, $vault)
{
    $k = random_bytes(16);
    $vault->rotate($k);
    return sodium_crypto_secretbox($m, $n, $k);
}

function unknownFunctionArgument($m, $n)
{
    $k = random_bytes(16);
    rotate_key_somewhere($k);
    return sodium_crypto_secretbox($m, $n, $k);
}

$topLevel = random_bytes(16);
$ref = &$topLevel;
$ref = random_bytes(32);
sodium_crypto_secretbox($message, $nonce, $topLevel);
