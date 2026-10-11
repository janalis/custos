<?php
<error descr="Supply a correctly sized secretbox key.">sodium_crypto_secretbox($message, $nonce, random_bytes(16))</error>;

function keptAfterPlainReads($m, $n, $other, $rows)
{
    $k = random_bytes(16);
    $copy = $k;
    $size = strlen($k) + count([$k]);
    $pair = [$k, 'label' => $k];
    [$k => $picked] = $rows;
    foreach ([$k] as $item) {
        echo $item;
    }
    $alias = &$other;
    $helper = function () use ($k, &$other) {
        return $k . $other;
    };
    $arrow = fn () => $k;
    return <error descr="Supply a correctly sized secretbox key.">sodium_crypto_secretbox($m, $n, $k)</error>;
}

function keptForStaticKey($m, $n)
{
    static $k = 'x';
    $k = random_bytes(16);
    return <error descr="Supply a correctly sized secretbox key.">sodium_crypto_secretbox_open($m, $n, $k)</error>;
}

function keptWhenWrittenAfterwards($m, $n)
{
    $k = str_repeat('a', 8);
    $box = <error descr="Supply a correctly sized secretbox key.">sodium_crypto_secretbox($m, $n, ($k))</error>;
    [$k] = [random_bytes(32)];
    return $box;
}
