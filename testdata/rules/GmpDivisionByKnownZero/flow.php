<?php
function known(): GMP
{
    $zero = 0;
    $divisor = $zero;
    return <error descr="Use a nonzero GMP divisor.">gmp_div_q(10, $divisor)</error>;
}

// A by-value capture cannot change the local.
function captured(): GMP
{
    $n = '0';
    $read = function () use ($n) {
        return extract($n);
    };
    return <error descr="Use a nonzero GMP divisor.">gmp_mod(10, $n)</error>;
}
