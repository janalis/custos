<?php
function blocking_alone($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (<warning descr="Require a completed TLS handshake before writing.">stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)</warning> !== false) {
        stream_set_blocking($s, true);
        fwrite($s, $secret);
        return true;
    }
    return false;
}

function write_before_guard($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (<warning descr="Require a completed TLS handshake before writing.">stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)</warning> !== false) {
        fwrite($s, $secret);
        if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== true) {
            return false;
        }
        return true;
    }
    return false;
}

function guard_accepts_pending($s, string $secret): bool
{
    if (!stream_set_blocking($s, false)) { return false; }
    if (<warning descr="Require a completed TLS handshake before writing.">stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)</warning> !== false) {
        if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) === false) {
            return false;
        }
        fwrite($s, $secret);
        return true;
    }
    return false;
}

function unbraced_write($s, string $secret): void
{
    if (!stream_set_blocking($s, false)) { return; }
    if (<warning descr="Require a completed TLS handshake before writing.">stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)</warning> !== false)
        fwrite($s, $secret);
}
