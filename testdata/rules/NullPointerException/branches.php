<?php
class Twig
{
}

class Leaf
{
    public $value;
    public $p;
    public function next(): ?Leaf { return null; }
    public function take(Leaf $l) {}
    public function m() { return $this; }

    public function arrows()
    {
        return fn (?Leaf $n) => $this->next()<warning descr="Possible null dereference.">-></warning>next();
    }

    public function tests($m, $x)
    {
        $y = $this->next() ?? $x;
        if ($this->next() instanceof Leaf) {
            $this->next()->next();
        }
        if (($this->next())) {
        } elseif ($this->m()->next()) {
            $this->m()->next()->next();
        }
        while ($this->m()->m()->next()) {
            $this->m()->m()->next()->next();
        }
        do {
        } while ($this->m()->m()->m()->next());
        $this->m()->m()->m()->next()->next();
        $z = !$x->next();
        $x->next()->next();
        $w = ($this->next()) || $x;
        $a?->$m()->next();
        $l = new Leaf();
        $l?->missing()->next();
    }

    public function unions(Leaf|Twig $z, ?Leaf $a, ?Leaf $b, ?Leaf $c, ?Leaf $d, $x)
    {
        $z?->next()->next();
        $q = $x && ($a);
        $a->value = 1;
        $v = !is_object($b) ? 0 : $b->value;
        $w = is_null(($c)) ?: $c->value;
        if (isset(<warning descr="Possible null dereference.">$d</warning>['k'])) {
            $d->value = 1;
        }
        if ($x) {
        } else {
            <warning descr="Possible null dereference.">$d</warning>->value = 2;
        }
    }

    public function locals($xs, $m)
    {
        $n = $this->next();
        if ($n = $this->next()) {
        }
        /** plain comment */
        $o = $this->next();
        <warning descr="Possible null dereference.">$o</warning>->value = 1;
        /** @var $q Leaf */
        $q = $this->next();
        $q->value = 1;
        /** @var $other Leaf */
        $r = $this->next();
        <warning descr="Possible null dereference.">$r</warning>->value = 1;
        /** @var Leaf $other */
        $s = $this->next();
        <warning descr="Possible null dereference.">$s</warning>->value = 1;
        $cb = function () use ($n) { return $n->value; };
        $this->$m($n);
        unknown_fn($n);
        $this->take($xs, $n);
        $this->that($n)->foo();
        $this->wrap(Assert::that($n));
        $t = $this->next();
        try {
        } catch (Exception $t) {
        }
        $t->value = 1;
    }

    public function chains(?Leaf $a, ?Leaf $b)
    {
        echo $a->p->value ?? 0;
        <warning descr="Possible null dereference.">$b</warning>->p->value = 1;
    }

    public function guards(?Leaf $a, ?Leaf $b, ?Leaf $c, ?Leaf $d, ?Leaf $e, ?Leaf $f, ?Leaf $g, ?Leaf $h, $x, $y)
    {
        if ($a?->p === null) {
        } elseif ($x) {
            $a->value = 1;
        }
        if ($x) {
        } elseif ($y) {
        } elseif (($b) instanceof Leaf) {
            $b->value = 1;
        }
        if (empty($c?->p)) {
        } elseif ($x) {
        } else {
            $c->value = 1;
        }
        $v = null === $d?->p || $d->value;
        if (is_object($e) && $x) {
            $e->value = 1;
        }
        if (is_null(($f))) {
        } else {
            $f->value = 1;
        }
        if ($g?->m()) {
            $g->value = 1;
        }
        if (isset($h?->p)) {
            $h->value = 1;
        }
    }

    public function unproven(?Leaf $a, ?Leaf $b, ?Leaf $c, ?Leaf $d, ?Leaf $e, ?Leaf $f, ?Leaf $g, $x, $y, $u)
    {
        if (f() !== null) {
            <warning descr="Possible null dereference.">$a</warning>->value = 1;
        }
        if (count($x)) {
            <warning descr="Possible null dereference.">$b</warning>->value = 1;
        }
        if (null !== $c?->p) {
            $c->value = 1;
        }
        if ($x === $y) {
            <warning descr="Possible null dereference.">$d</warning>->value = 1;
        }
        if ($e?->p === $u) {
            <warning descr="Possible null dereference.">$e</warning>->value = 1;
        }
        if (true) {
            <warning descr="Possible null dereference.">$f</warning>->value = 1;
        }
        if (is_object($g)) {
            foreach ($x as $g) {
            }
            <warning descr="Possible null dereference.">$g</warning>->value = 1;
        }
    }
}
