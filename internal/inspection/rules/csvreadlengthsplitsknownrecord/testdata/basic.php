<?php
$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); <warning descr="Allow enough bytes for the complete CSV record.">fgetcsv($h, 4)</warning>;

function untouched($other)
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    stream_get_line($other, 6);

    return <warning descr="Allow enough bytes for the complete CSV record.">fgetcsv($h, 4)</warning>;
}
