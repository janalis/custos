<?php
function power_of_pair()
{
    $pair = gmp_div_qr(10, 3);
    return <error descr="Select a quotient or remainder before using the GMP result.">gmp_pow($pair, 2)</error>;
}

function int_of_pair()
{
    return <error descr="Select a quotient or remainder before using the GMP result.">gmp_intval(gmp_div_qr(10, 3))</error>;
}

function power_of_quotient()
{
    $pair = gmp_div_qr(10, 3);
    return gmp_pow($pair[0], 2);
}
