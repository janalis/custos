<?php
class Row
{
    public $id;

    public function __get($k) { return null; }

    public function fill(): void
    {
        $this->label = 'x';
    }
}

function label(Row $r, Row $other): bool
{
    $other->extra = true;
    return isset($r->extra) || isset(<error descr="\Row has no __isset(); this isset/empty check is always false.">$r->label</error>);
}
