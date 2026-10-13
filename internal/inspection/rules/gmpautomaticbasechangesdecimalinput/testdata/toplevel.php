<?php

function untouched()
{
    global $elsewhere;

    return $elsewhere;
}

$local = '024';

echo <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init($local)</warning>;
