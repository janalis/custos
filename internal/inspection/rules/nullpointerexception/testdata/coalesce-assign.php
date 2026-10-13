<?php
class Box
{
    public $size;
    private Box $fallback;

    public function next(): ?Box { return null; }

    public function open(?Box $b = null, ?Box $c = null): void
    {
        $b ??= $this->fallback;
        $b->size = 1;
        $c ??= $this->next();
        <warning descr="Possible null dereference.">$c</warning>->size = 2;
    }
}
