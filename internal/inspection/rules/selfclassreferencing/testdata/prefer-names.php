<?php
class Ledger
{
    public static function open(<weak_warning descr="Spell the class reference &apos;self&apos; as &apos;Ledger&apos;.">self</weak_warning> $base): <weak_warning descr="Spell the class reference &apos;self&apos; as &apos;Ledger&apos;.">self</weak_warning>
    {
        tag(<weak_warning descr="Spell the class reference &apos;__CLASS__&apos; as &apos;Ledger::class&apos;.">__CLASS__</weak_warning>);
        tag(<weak_warning descr="Spell the class reference &apos;self&apos; as &apos;Ledger&apos;.">self</weak_warning>::class);
        $f = fn() => new self;
        tag(<weak_warning descr="Spell the class reference &apos;SELF&apos; as &apos;Ledger&apos;.">SELF</weak_warning>::X, <weak_warning descr="Spell the class reference &apos;__class__&apos; as &apos;Ledger::class&apos;.">__class__</weak_warning>, static::class);
        return new <weak_warning descr="Spell the class reference &apos;self&apos; as &apos;Ledger&apos;.">self</weak_warning>;
    }

    public function plain() {
        return new Ledger;
    }
}

__CLASS__;
