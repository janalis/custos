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
            <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning> $this->rate,
            <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $this->code,
            (float) ($cell * 100),
            (int) ($cell * 100),
            (float) ($n * 2),
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> ($n * 2),
            <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning> ($cell * $f),
            (int) ($n ** $n),
            (float) ($unknown * 2),
        ];
    }
}
