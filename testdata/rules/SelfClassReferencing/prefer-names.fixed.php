<?php
class Ledger
{
    public static function open(Ledger $base): Ledger
    {
        tag(Ledger::class);
        tag(Ledger::class);
        $f = fn() => new self;
        tag(Ledger::X, Ledger::class, static::class);
        return new Ledger;
    }

    public function plain() {
        return new Ledger;
    }
}

__CLASS__;
