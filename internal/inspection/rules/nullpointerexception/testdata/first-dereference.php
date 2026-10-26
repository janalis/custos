<?php
class Box
{
    public ?Box $next = null;
    public function open(): ?Box { return null; }
}

class Opener
{
    public function take(Box $b): void {}

    public function params(?Box $b, ?Box $c, ?Box $d): void
    {
        <warning descr="Possible null dereference.">$b</warning>->open();
        $b->open();
        $this->take(<warning descr="Possible null dereference.">$c</warning>);
        $c->open();
        echo <warning descr="Possible null dereference.">$d</warning>->next;
        echo $d->next['k'] ?? 1;
        $z = isset($d->next['k'], $b);
        $this->take($b, $d);
        <warning descr="Possible null dereference.">$d</warning>->open();
    }

    public function locals(Box $box, bool $flag): void
    {
        $a = $box->open();
        if ($flag) {
            <warning descr="Possible null dereference.">$a</warning>->open();
        }
        <warning descr="Possible null dereference.">$a</warning>->open();
        $a->open();

        $b = $box->open();
        $flag && <warning descr="Possible null dereference.">$b</warning>->open();
        $x = $flag ? <warning descr="Possible null dereference.">$b</warning>->open() : null;
        return;
    }

    public function nested(Box $box, ?Box $other): ?Box
    {
        $a = $box->open();
        $other?->open(<warning descr="Possible null dereference.">$a</warning>->open());
        $b = $box->open();
        return (<warning descr="Possible null dereference.">$b</warning>->open() ?? $b)->open();
    }

    public function condition(Box $box): void
    {
        if ($a = $box->open()) {
        }
        $e = $box->open();
        $box?->{<warning descr="Possible null dereference.">$e</warning>->open()};
        <warning descr="Possible null dereference.">$e</warning>->open();
        $c = $box->open();
        while (<warning descr="Possible null dereference.">$c</warning>->open()) {
            $c->open();
        }
    }
}
