<?php

trait Stamps
{
    public function stamp() { return 'trait'; }
}

trait Seals
{
    public function seal() { return 'seal'; }
}

class Ledger
{
    public function total()
    {
        return static::rate()
            . static::stamp()
            . static::mark()
            . $this->round()
            . $this->label()
            . self::title()
            . $this->label();
    }

    private function rate() { return 'ledger'; }
    private function stamp() { return 'ledger'; }
    private function mark() { return 'ledger'; }
    private function round() { return 'ledger'; }
    public function label() { return 'ledger'; }
    public function title() { return 'ledger'; }
}

class TaxLedger extends Ledger
{
    use Stamps;
    use Seals { seal as mark; seal as title; }

    public function rate() { return 'tax'; }
}

final class Vault
{
    public function open() { return $this->code(); }
    private function code() { return 1; }
}

trait Counts
{
    public function count() { return static::step() + $this->base(); }
    private function step() { return 1; }
    public function base() { return 0; }
}

class LoopA extends LoopB
{
    public function run() { return static::pick(); }
    private function pick() { return 1; }
}

class LoopB extends LoopA
{
}
