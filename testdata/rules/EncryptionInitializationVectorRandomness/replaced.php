<?php
// Placeholders that a guard, try or do-while always replaces before the call.
function sealDefault(string $data, string $key, ?string $iv = null, $other = 1): string
{
    if ($iv === null) {
        $iv = random_bytes(16);
    }
    return openssl_encrypt($data, 'aes-256-cbc', $key, OPENSSL_RAW_DATA, $iv);
}

function sealYoda($data, $key, $iv = null)
{
    if (null == $iv)
        $iv = random_bytes(16);
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

function sealIsNull($data, $key, $iv = null)
{
    if (is_null($iv)) {
        throw new \InvalidArgumentException('IV required');
    }
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

function sealFalsy($data, $key, $mode)
{
    $iv = '';
    if (!$iv) {
        $iv = random_bytes(16);
    }
    $nonce = null;
    switch ($mode) {
        case 'gcm':
            if (empty($nonce)) {
                $nonce = random_bytes(12);
            }
            return openssl_encrypt($data, 'aes-256-gcm', $key, 0, $nonce);
    }
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

function sealTry($data, $key)
{
    $iv = null;
    try {
        $iv = random_bytes(16);
    } catch (\Error $e) {
        $iv = openssl_random_pseudo_bytes(16);
    } catch (\Exception $e) {
        throw new \RuntimeException('no entropy', 0, $e);
    }
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

function sealRetry($data, $key)
{
    $iv = '';
    do {
        $iv = random_bytes(16);
    } while ($iv[0] === "\0");
    return openssl_encrypt($data, 'aes-256-cbc', $key, 0, $iv);
}

// Still reported: the placeholder can reach the call.
function weakGuards($data, $key, $flag, $iv = null)
{
    if ($iv !== null) {
        $iv = random_bytes(16);
    }
    $a = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$iv</error>);

    $b = '';
    if ($b === null) {
        $b = random_bytes(16);
    }
    $b2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: ''.">$b</error>);

    $c = null;
    if ($flag) {
        $c = random_bytes(16);
    }
    $c2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$c</error>);

    $d = null;
    if ($d === null) {
        log_it('generating');
    }
    $d2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$d</error>);

    $e = 'seed';
    if (!$e) {
        $e = random_bytes(16);
    }
    $e2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'seed'.">$e</error>);
    return [$a, $b2, $c2, $d2, $e2];
}

function weakBlocks($data, $key, $flag)
{
    $a = null;
    try {
        log_it('start');
    } catch (\Exception $e) {
        throw $e;
    }
    $a2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$a</error>);

    $b = null;
    try {
        $b = random_bytes(16);
    } catch (\Exception $e) {
        log_it('fallback');
    }
    $b2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$b</error>);

    $c = '';
    do {
        if ($flag) {
            break;
        }
        $c = random_bytes(16);
    } while (false);
    $c2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: ''.">$c</error>);

    if ($d === null) {
        $d = random_bytes(16);
    }
    if ($flag) {
        $d = null;
    }
    $d2 = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: null.">$d</error>);
    return [$a2, $b2, $c2, $d2];
}
