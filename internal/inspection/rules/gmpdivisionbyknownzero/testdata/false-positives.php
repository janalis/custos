<?php
// A divisor counted with ++ is not a known zero.
function average(array $values): string
{
    $sum = gmp_init(0);
    $count = 0;
    foreach ($values as $v) {
        $sum = gmp_add($sum, $v);
        $count++;
    }
    return gmp_strval(gmp_div_q($sum, $count));
}

function countdown(): GMP
{
    $step = 0;
    --$step;
    return gmp_mod(10, $step);
}

// The increment reaches the divisor through a copy.
function copied(array $values): GMP
{
    $seen = 0;
    foreach ($values as $v) {
        $seen++;
    }
    $divisor = gmp_init($seen);
    return gmp_div_r(10, $divisor);
}

function caught(): GMP
{
    $e = 0;
    try {
        risky();
    } catch (Exception $e) {
    }
    return gmp_div_q(10, $e);
}

function byReference(): GMP
{
    $n = 0;
    $bump = function () use (&$n) {
        $n = 4;
    };
    $bump();
    return gmp_div_q(10, $n);
}

function imported(array $row): GMP
{
    $n = 0;
    extract($row);
    return gmp_div_q(10, $n);
}
