<?php
function strict_guard($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) {
        stream_set_blocking($s, true);
        if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== true) {
            return false;
        }
        fwrite($s, $secret);
        return true;
    }
    return false;
}

function negated_guard($s, string $secret): void
{
    if (!stream_set_blocking($s, false)) { return; }
    if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) {
        if (!stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)) {
            throw new RuntimeException('TLS handshake incomplete');
        }
        fwrite($s, $secret);
    }
}

function assigned_guard($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) {
        if (($done = stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)) != true) {
            return (bool) $done;
        }
        fwrite($s, $secret);
        return true;
    }
    return false;
}
