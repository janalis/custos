<?php
namespace Shop;

interface Countable2 {}

class Crate
{
    public function pair(Crate&Countable2 $other): Crate&Countable2
    {
        return $other;
    }

    public function maybe((Crate&Countable2)|null $other): (Crate&Countable2)|self
    {
        return $other ?? $this;
    }
}
