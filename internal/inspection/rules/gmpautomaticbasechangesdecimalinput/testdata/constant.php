<?php

final class Ledger
{
    const DIGITS = '017';

    public function total()
    {
        return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init(self::DIGITS)</warning>;
    }
}
