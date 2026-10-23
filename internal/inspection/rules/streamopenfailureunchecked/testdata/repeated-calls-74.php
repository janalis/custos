<?php
// Before PHP 8 a stream function only warns about false: every call counts.
function load(string $path): string
{
    $stream = fopen($path, 'rb');
    $magic = <warning descr="Check that the stream opened successfully.">fread($stream, 4)</warning>;
    return $magic . <warning descr="Check that the stream opened successfully.">fread($stream, 64)</warning>;
}
