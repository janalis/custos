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
        return <warning descr="Call 'rate' with '$this->' instead of '::'.">static::rate()</warning>
            . <warning descr="Call 'stamp' with '$this->' instead of '::'.">static::stamp()</warning>
            . <warning descr="Call 'mark' with '$this->' instead of '::'.">static::mark()</warning>
            . <warning descr="Call 'round' with '$this->' instead of '::'.">static::round()</warning>
            . <warning descr="Call 'label' with '$this->' instead of '::'.">static::label()</warning>
            . <warning descr="Call 'title' with '$this->' instead of '::'.">self::title()</warning>
            . <warning descr="Call 'label' with '$this->' instead of '::'.">self::label()</warning>;
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
    public function open() { return <warning descr="Call 'code' with '$this->' instead of '::'.">static::code()</warning>; }
    private function code() { return 1; }
}

trait Counts
{
    public function count() { return <warning descr="Call 'step' with '$this->' instead of '::'.">static::step()</warning> + <warning descr="Call 'base' with '$this->' instead of '::'.">static::base()</warning>; }
    private function step() { return 1; }
    public function base() { return 0; }
}

class LoopA extends LoopB
{
    public function run() { return <warning descr="Call 'pick' with '$this->' instead of '::'.">static::pick()</warning>; }
    private function pick() { return 1; }
}

class LoopB extends LoopA
{
}
