<?php
class Ledger
{
    private $rows;

    public function <error descr="Class 'Ledger' uses an old-style constructor; rename it to __construct.">Ledger</error>(array $rows)
    {
        $this->rows = $rows;
    }
}

abstract class Shape
{
    abstract protected function <error descr="Class 'Shape' uses an old-style constructor; rename it to __construct.">Shape</error>($sides);
}

final class Tally
{
    function <error descr="Class 'Tally' uses an old-style constructor; rename it to __construct.">Tally</error>() {}
}
