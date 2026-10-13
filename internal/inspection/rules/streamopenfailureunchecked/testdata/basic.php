<?php
function bad($path) { $h = fopen($path, 'rb'); <warning descr="Check that the stream opened successfully.">fread($h, 12)</warning>; }

function logsOnly(string $path): void
{
    $h = fopen($path, 'rb');
    if (!$h) {
        echo 'cannot open';
    }
    <warning descr="Check that the stream opened successfully.">fread($h, 4)</warning>;
}

function wrongBranch(string $path)
{
    $h = fopen($path, 'rb');
    return $h ? null : <warning descr="Check that the stream opened successfully.">fgets($h)</warning>;
}

function guardThenReopen(string $path, string $other): void
{
    $h = fopen($path, 'rb');
    if (!$h) {
        return;
    }
    $h = fopen($other, 'rb');
    <warning descr="Check that the stream opened successfully.">fread($h, 4)</warning>;
}

function assertion(string $path): void
{
    $h = fopen($path, 'rb');
    assert($h !== false);
    <warning descr="Check that the stream opened successfully.">rewind($h)</warning>;
}

function otherVariable(string $path, $ready): void
{
    $h = fopen($path, 'rb');
    if ($ready) {
        <warning descr="Check that the stream opened successfully.">fwrite($h, 'x')</warning>;
    }
}
