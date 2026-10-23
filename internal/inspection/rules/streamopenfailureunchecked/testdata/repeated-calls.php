<?php
// On PHP 8 a stream function throws a TypeError for false: once one call
// on the handle returned, it is open. The first call is the one to guard.
function load(string $path): string
{
    $stream = fopen($path, 'rb');
    $magic = <warning descr="Check that the stream opened successfully.">fread($stream, 4)</warning>;
    fseek($stream, 8);
    $body = fread($stream, 64);
    return $magic . $body;
}

// Inside a try with a catch, the TypeError may be caught: each call counts.
function guarded(string $path): string
{
    try {
        $stream = fopen($path, 'rb');
        $magic = <warning descr="Check that the stream opened successfully.">fread($stream, 4)</warning>;
        return $magic . <warning descr="Check that the stream opened successfully.">fread($stream, 4)</warning>;
    } catch (\TypeError) {
        return '';
    }
}
