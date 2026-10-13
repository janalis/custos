<?php
class Ledger
{
    private $rows;

    public function __construct(array $rows)
    {
        $this->rows = $rows;
    }
}

abstract class Shape
{
    abstract protected function __construct($sides);
}

final class Tally
{
    function __construct() {}
}
