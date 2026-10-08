<?php

final class Ledger
{
    /** @var float */
    private $cost;

    /** @var float */
    private $rate;

    /** @var string */
    private $code = 'n/a';

    public function __construct()
    {
        $this->rate = 1.0;
    }

    /** @param string|null $cell */
    public function amounts($cell, int $n, float $f)
    {
        return [
            (float) $this->cost,
            (float) $this->rate,
            (string) $this->code,
            (float) ($cell * 100),
            (int) ($cell * 100),
            (float) ($n * 2),
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> ($n * 2),
            (float) ($cell * $f),
            (int) ($n ** $n),
            (float) ($unknown * 2),
        ];
    }
}
