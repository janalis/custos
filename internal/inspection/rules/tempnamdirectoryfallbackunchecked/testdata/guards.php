<?php

function recreatedAfterGuard(string $privateDir, string $secret): void
{
    $path = tempnam($privateDir, 'job');
    if (realpath(dirname($path)) !== realpath($privateDir)) {
        return;
    }
    $path = tempnam($privateDir, 'next');
    <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>;
}

function uncanonicalGuard(string $privateDir, string $secret): void
{
    $path = tempnam($privateDir, 'job');
    if (dirname($path) !== $privateDir) {
        return;
    }
    <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>;
}

function matchDoesNotCoverElse(string $privateDir, string $secret): void
{
    $path = tempnam($privateDir, 'job');
    if (realpath(dirname($path)) === realpath($privateDir)) {
        file_put_contents($path, $secret);
    } else {
        <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>;
    }
}

function systemSubfolder(string $secret)
{
    $path = tempnam(sys_get_temp_dir() . '/private', 'job');
    return <warning descr="Verify the temporary file remains in the required directory.">fopen($path, 'w')</warning>;
}
