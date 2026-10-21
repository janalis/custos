<?php
class Shelf
{
    public function top(): ?Shelf { return null; }
    public function tidy(): void {}
}

function lookup(): ?Shelf { return null; }

function guarded(Shelf $s, $item): void
{
    if ($item instanceof Shelf && null !== $s->top()) {
        $top = $s->top();
        $top->tidy();
    }
}

function earlyExit(): void
{
    if (lookup() === null) {
        return;
    }
    $found = lookup();
    $found->tidy();
    if (!lookup()) {
        return;
    }
    $other = lookup();
    $other->tidy();
}

function testedElsewhere(Shelf $s, Shelf $t): void
{
    if (null !== $t->top()) {
        $t->tidy();
    }
    $top = $s->top();
    <warning descr="Possible null dereference.">$top</warning>->tidy();
    $local = $s->top();
    if (null !== $s->top()) {
        <warning descr="Possible null dereference.">$local</warning>->tidy();
    }
    $plain = $s;
    $plain->tidy();
}
