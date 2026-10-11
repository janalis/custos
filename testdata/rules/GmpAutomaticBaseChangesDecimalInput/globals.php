<?php

function reset_digits()
{
    global $shared, $$name;
    $shared = '0x1F';
}

function reset_other()
{
    $GLOBALS['viaSuperglobal'] = '0x1F';
}

$shared = '010';
$viaSuperglobal = '010';
$local = '024';
reset_digits();
reset_other();

echo gmp_init($shared);
echo gmp_init($viaSuperglobal);
