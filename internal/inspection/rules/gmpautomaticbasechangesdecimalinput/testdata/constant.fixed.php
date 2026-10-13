<?php

final class Ledger
{
    const DIGITS = '017';

    public function total()
    {
        return gmp_init(self::DIGITS, 10);
    }
}
