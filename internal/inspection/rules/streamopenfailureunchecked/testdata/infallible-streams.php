<?php
// PHP's own memory and standard streams always open.
function buffers(string $data): string
{
    $memory = fopen('php://memory', 'r+');
    fwrite($memory, $data);
    $temp = fopen('php://temp/maxmemory:1048576', 'r+');
    fwrite($temp, $data);
    $upper = fopen('PHP://TEMP', 'r+');
    rewind($upper);
    $out = fopen("php://output", 'w');
    fwrite($out, $data);
    $err = fopen('php://stderr', 'w');
    fflush($err);
    return (string) ftell($memory);
}

// A file descriptor, a filter or a file may not open.
function others(string $path): void
{
    $fd = fopen('php://fd/3', 'w');
    <warning descr="Check that the stream opened successfully.">fwrite($fd, 'x')</warning>;
    $filtered = fopen('php://filter/read=string.rot13/resource=' . $path, 'r');
    <warning descr="Check that the stream opened successfully.">fread($filtered, 8)</warning>;
    $named = fopen($path, 'r');
    <warning descr="Check that the stream opened successfully.">fread($named, 8)</warning>;
    $number = fopen(0, 'r');
    <warning descr="Check that the stream opened successfully.">fread($number, 8)</warning>;
}
