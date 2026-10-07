<?php
namespace Storage;

function fopen($path, $mode) { return null; }

function open_files($path)
{
    $own = fopen($path, 'r');
    $h1 = \FOpen($path, 'rb');
}
