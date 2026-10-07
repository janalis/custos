<?php
function open_files($path)
{
    $access = 'ab';
    $h1 = fopen($path, $access);
    $h2 = fopen($path, 'rb+');
    $h3 = fopen($path, 'ab');
    $h4 = \fopen($path, 'xb');
    $h5 = fopen($path, 'wb+');
    $h6 = fopen($path, 'rb');
}
