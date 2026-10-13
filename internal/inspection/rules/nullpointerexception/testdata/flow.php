<?php
class Item
{
    public $value;
    public $name;
    public function next(): ?Item { return null; }
    public function same(Item $other): bool { return true; }

    public function branches(?Item $a, ?Item $b, ?Item $c, ?Item $d, ?Item $e, Item $known, ?Item $f, ?Item $g)
    {
        if (is_object($a)) {
            $a->value = 1;
        }
        <warning descr="Possible null dereference.">$a</warning>->value = 2;

        if ($b === $known) {
            $b->value = 1;
        }
        $x = $c?->value !== null ? $c->value : 0;
        $y = $d?->next() && $d->next();
        if (!is_object($e)) {
            echo 'none';
        } else {
            $e->value = 1;
        }
        while (is_object($f)) {
            $f = $f->next();
        }
        if (is_object($g)) {
            $g = $this->next();
            <warning descr="Possible null dereference.">$g</warning>->value = 1;
        }
    }

    public function cased(?Item $a, ?Item $b, ?Item $c)
    {
        self::AssertNotNull($a);
        Verify::That($b)->NOTNULL();
        IS_NULL($c);
        return [$a->value, $b->value, $c->value];
    }

    public function loose(?Item $a, ?Item $b, string $s)
    {
        if ($a == $s) {
            <warning descr="Possible null dereference.">$a</warning>->value = 1;
        }
        if (is_object($b) || $s) {
            <warning descr="Possible null dereference.">$b</warning>->value = 1;
        }
    }

    public function chainCompared(?Item $a, ?Item $b, string $s)
    {
        if ('v' === <warning descr="Possible null dereference.">$a</warning>->name) {
            return $a->value;
        }
        if (<warning descr="Possible null dereference.">$b</warning>->name == $s) {
            return <warning descr="Possible null dereference.">$b</warning>->value;
        }
        return 0;
    }
}
