<?php
function archive(string $path, string $stub)
{
    $fp = fopen($path, 'rb');
    <warning descr="Check that the stream opened successfully.">fseek($fp, 4)</warning>;
    $head = fread($fp, 4);
    include $stub;
    return $head;
}

function reread(string $path, string $stub)
{
    $fp = fopen($path, 'rb');
    <warning descr="Check that the stream opened successfully.">fseek($fp, 4)</warning>;
    include $stub;
    return fread($fp, 4);
}
