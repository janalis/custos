<?php
namespace Storage;

function fopen($path, $mode) { return null; }

function open_files($path)
{
    $own = fopen($path, 'r');
    $h1 = \FOpen($path, <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'br'</error>);
}
