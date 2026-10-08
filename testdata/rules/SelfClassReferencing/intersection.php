<?php
namespace Shop;

interface Countable2 {}

class Crate
{
    public function pair(Crate&Countable2 $other): Crate&Countable2
    {
        return $other;
    }

    public function maybe((Crate&Countable2)|null $other): (Crate&Countable2)|<weak_warning descr="Refer to the class as 'self' instead of 'Crate'.">Crate</weak_warning>
    {
        return $other ?? $this;
    }
}
