<?php
$h = fopen('data://text/plain,abcdefgh,ij%0A', 'r'); fgetcsv($h, 100);

function tail()
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    stream_get_line($h, 6);

    return fgetcsv($h, 4, ",", "\"", "\\");
}

function copied($out)
{
    $h = fopen('data://text/plain,abcdefgh', 'r');
    stream_copy_to_stream($h, $out, 6);

    return fgetcsv($h, 4);
}
