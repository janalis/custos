<?php
function open_files($path, $flag)
{
    $either = $flag ? 'r' : 'w';
    $h5 = fopen($path, 'r+b');
    $h6 = fopen($path, 'xtb');
    $h7 = fopen($path, '');
    $h8 = fopen($path);
    $h9 = fopen($path, $_GET['m']);
    $h10 = fopen($path, $either);
    $h12 = fopen($path, 'b');
}
