<?php
// E7: the fallback that runs once random_bytes() is known to be missing.
function token($n)
{
    if (function_exists('random_bytes')) {
        return random_bytes($n);
    } elseif (\function_exists('\random_bytes') && $n > 4) {
        return '';
    } elseif ($n > 8) {
        return openssl_random_pseudo_bytes($n, $strong);
    } else {
        return openssl_random_pseudo_bytes($n, $strong);
    }
}

function legacy($n)
{
    if (!function_exists('random_bytes')) {
        return openssl_random_pseudo_bytes($n, $strong);
    }
    if (!function_exists('random_int')) {
        return <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong);
    }
    if (function_exists('random_bytes')) {
        return <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong);
    } elseif (function_exists($n)) {
        $f = function () use ($n) {
            return <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong);
        };
    }
    if (<weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong)) {
    } elseif (function_exists('random_bytes', 1)) {
    } else {
        return <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong);
    }
    return '';
}

function later($n, $fast)
{
    if ($fast) {
        return '';
    } elseif (function_exists($n)) {
        return '';
    } elseif (function_exists('random_bytes')) {
        return random_bytes($n);
    } else {
        return openssl_random_pseudo_bytes($n, $strong);
    }
}

function before($n, $fast)
{
    if ($fast) {
        return '';
    } elseif ($n) {
        return <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>($n, $strong);
    } elseif (function_exists('random_bytes')) {
        return random_bytes($n);
    }
    return '';
}
